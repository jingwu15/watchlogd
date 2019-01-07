package watch

import (
	"os"
	"time"
	"bytes"
	"syscall"
	"strconv"
	"strings"
	"os/exec"
	"io/ioutil"
	"os/signal"
	"path/filepath"
	"github.com/spf13/viper"
	//"github.com/erikdubbelboer/gspt"
	log "github.com/sirupsen/logrus"
	logchan "github.com/jingwu15/golib/logchan"
)

var runFlag int = 1
func findProcess(procTitle string) ([]int, error) {
	var err error
	matches, err := filepath.Glob("/proc/*/cmdline")
	if err != nil {
		return nil, err
	}
	var pid int
	var pids = []int{}
	var tmp []string
	var body []byte
	for _, filename := range matches {
		body, err = ioutil.ReadFile(filename)
		if err == nil {
			//if bytes.HasPrefix(body, []byte(procTitle)) {
			if bytes.Contains(body, []byte(procTitle)) {
				tmp = strings.Split(filename, "/")
				pid, _ = strconv.Atoi(tmp[2])
				pids = append(pids, pid)
			}
		}
	}
	return pids, nil
}

//初始化日志
func InitLog() {
	go logchan.LogWrite()
	log.SetFormatter(&log.JSONFormatter{})
	//log.SetOutput(ioutil.Discard)
	log.SetLevel(log.DebugLevel)
	config := map[string]string{
		"error":      viper.GetString("log_error"),
		"info":       viper.GetString("log_info"),
		"writeDelay": "1",
		"cutType":    "day",
	}
	logChanHook := logchan.NewLogChanHook(config)
	log.AddHook(&logChanHook)
}

//处理信号，停止及重启
func handleSignals() {
	var sig os.Signal
	var signalChan = make(chan os.Signal, 100)
	signal.Notify(
		signalChan,
		syscall.SIGTERM,
		syscall.SIGUSR2,
	)
	for {
		sig = <-signalChan
		switch sig {
		case syscall.SIGTERM:
			log.Info("stop")
			logchan.LogClose()
            runFlag = 0
		case syscall.SIGUSR2:
			log.Info("restart")
			logchan.LogClose()
		default:
		}
	}
}

func Run() {
	procTitle := viper.GetString("proc_title")
	//gspt.SetProcTitle(procTitle)

	go handleSignals()
	InitLog()

    //执行
	go DoWatch()
    log.Info(procTitle + " is running")
    //done := make(chan bool)
    //<-done
    for {
        if runFlag == 1 {
            //未结束，一直等待
            time.Sleep(time.Duration(2)*time.Second)
            //log.Info(procTitle + " is running")
        } else {
            log.Info(procTitle + " is shut down")
            break;
        }
    }
}

func Start() {
	var err error
	procTitle := viper.GetString("proc_title")
	cmdFile, _ := filepath.Abs(os.Args[0])
	cmd := "nohup " + cmdFile + " run 2> /dev/null 1>/dev/null &"
	client := exec.Command("sh", "-c", cmd)
	err = client.Start()
	if err != nil {
		log.Info(procTitle + " start error:", err)
		return
	}
	err = client.Wait()
	if err != nil {
		log.Info(procTitle + " start error:", err)
		return
	}
	log.Info(procTitle + " is started")
	return
}

//func Restart() {
//	procTitle := viper.GetString("proc_title")
//	pids, err := findProcess(procTitle)
//	if err != nil {
//		log.Error(err)
//		return
//	}
//	for _, pid := range pids {
//		syscall.Kill(pid, syscall.SIGUSR2)
//	}
//	fmt.Println(procTitle + " is restarted")
//}

func Stop() {
	procTitle := viper.GetString("proc_title")
	pids, err := findProcess(procTitle)
	if err != nil {
		log.Error(err)
		return
	}
	for _, pid := range pids {
		syscall.Kill(pid, syscall.SIGTERM)
	}
	log.Info(procTitle + " is stoped")
}

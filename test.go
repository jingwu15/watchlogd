package main

import (
    "os"
    //"io"
    "fmt"
    //"bufio"
    //"bytes"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
    "github.com/jingwu15/watchlogd/watch"
	"github.com/jingwu15/golib/logchan"
	//"github.com/jingwu15/golib/beanstalk"
	//"github.com/jingwu15/golib/json"
)

func main() {
    //record := map[string]interface{}{}
    //fp, _ := os.Open("/tmp/log_phperror_sso-backend-server.log")
    //rd := bufio.NewReader(fp)
    //for {
    //    line, err := rd.ReadBytes('\n')    //以'\n'为结束符读入一行
    //    if err != nil || io.EOF == err {
    //        break
    //    }
    //    row := bytes.Split(line, []byte("\t"))

    //    fmt.Println(string(row[1]), "\n\n\n")
    //    err = json.Decode(row[1], &record)
    //    fmt.Println(err, record, "\n\n\n")
    //    record["syskey"] = "tester"
    //    res, err := json.Encode(record)
    //    fmt.Println(err, string(res), "\n\n\n")
    //    break
    //    //fmt.Println(line)
    //    //fmt.Println(row)
    //    //ctime, err := json.GetString([]byte(row[1]), "create_at")
    //    //fmt.Println(ctime, err)
    //    //fmt.Println(string(row[0]))
    //    //fmt.Println(string(row[1]))
    //}
    //return
    //id, err := beanstalk.Put("jw_tester", []byte(`jw_tester`))
    //fmt.Println(id, err)
    watch.DoWatch()
    done := make(chan bool)
    <-done
}

func init() {
	viper.SetConfigFile("./watchlogd.json")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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


package watch

import (
    "io"
    "os"
    //"log"
    "fmt"
    "path"
    "time"
    "bufio"
    "strings"
    //"unicode/utf8"
    "path/filepath"
    "github.com/spf13/viper"
    "github.com/hpcloud/tail"
    "github.com/fsnotify/fsnotify"
	"github.com/jingwu15/golib/json"
	log "github.com/sirupsen/logrus"
	//"github.com/jingwu15/golib/logchan"
	"github.com/jingwu15/golib/beanstalk"
)

type LogRecord struct {
    Queue string
    Body  []byte
}

var total = 0
var tailFiles = make(map[string]map[string]string)
var logQueue = make(chan LogRecord)

func ToQueue() {
    for {
        select{
        case record, ok := <- logQueue:
            if ok {
                _, err := beanstalk.Put(record.Queue, record.Body)
                if err != nil {
                    log.Error("beanstalk error: ", err)
                }
                total--
            }
        case <- time.After(time.Second * 10):
            //fmt.Println("ToQueue timeout")
        }
    }
}

func TailYpfError(queue string, syskey string, fpath string, timeout int) error {
    isTail := 1
    isWrite := 1
    isRemove :=  0
    record := map[string]interface{}{}
    finfo , err := os.Stat(fpath)
    if err != nil {
        return err
    }
    seek := finfo.Size()
    tailConfig := tail.Config{Follow: true, MustExist: false, Logger: tail.DiscardingLogger}

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    err = watcher.Add(fpath)
    if err != nil {
        return err
    }
    tailConfig.Location = &tail.SeekInfo{seek, os.SEEK_SET}
    tailHandle, err := tail.TailFile(fpath, tailConfig)
    for {
        if isRemove == 1 {
            break
        }
        if isTail == 0 && isWrite == 1 {
            tailConfig.Location = &tail.SeekInfo{seek, os.SEEK_SET}
            tailHandle, _ = tail.TailFile(fpath, tailConfig)
            isTail = 1
        }
        select {
        case line, ok := <- tailHandle.Lines:
            if ok {
                seek = seek + int64(len(line.Text)) + 1
                row := strings.Split(line.Text, "\t")
                //json
                err = json.Decode([]byte(row[1]), &record)
                if err == nil {
                    record["syskey"] = syskey
                    body, err := json.Encode(record)
                    if err == nil {
                        logQueue <- LogRecord{Queue:queue, Body:body}
                    }
                } else {
                    //非json 不处理
                }
                total++
            }

        case event, ok := <-watcher.Events:
            if ok && event.Op == fsnotify.Write {
                isWrite = 1
            }
            if ok && (event.Op == fsnotify.Remove || event.Op == fsnotify.Rename) {
                isRemove = 1       //删除/重命名
                watcher.Close()
                delete(tailFiles, fpath)
            }

        //case err, ok := <-watcher.Errors:
        //    fmt.Println("goTail watcher error", err, ok)

        case <- time.After(time.Second * time.Duration(timeout)):
            isTail = 0           //超时
            isWrite = 0
            tailHandle.Stop()
            //fmt.Println("goTail timeout")
        }
        if isTail == 0 && isWrite == 0 {
            time.Sleep(time.Duration(1)*time.Second)
        }
    }
    return nil
}

func ParsePhpError(raw string) map[string]string {
    tmp0 := strings.Split(raw[1:21], " ")
    ymd := strings.Split(tmp0[0], "-")
    createAt := fmt.Sprintf("%s-%s-%s %s", ymd[2], ymd[1], ymd[0], tmp0[1])
    return map[string]string{"create_at": createAt, "msg": raw}
}

//日志示例：
//[02-Jan-2019 23:58:19 Asia/Shanghai] PHP Fatal error:  Allowed memory size of 1073741824 bytes exhausted (tried to allocate 72 bytes) in /data/www/home-v3-cli/Service/Es/Admin/LogEsService.php on line 37
//[03-Jan-2019 15:16:17 Asia/Shanghai] PHP Fatal error:  Uncaught exception 'PDOException' with message 'SQLSTATE[HY000]: General error: 2006 MySQL server has gone away' in /usr/local/nginx-1.8.1/html/Home-V4-Cli/Ypf/Lib/DatabaseV5.php:98
//Stack trace:
//#0 /usr/local/nginx-1.8.1/html/Home-V4-Cli/Ypf/Lib/DatabaseV5.php(98): PDOStatement->execute(Array)
//#1 /usr/local/nginx-1.8.1/html/Home-V4-Cli/Ypf/Lib/DatabaseV5.php(343): Ypf\Lib\DatabaseV5->query('SELECT  *  FROM...', Array)
//#2 /usr/local/nginx-1.8.1/html/Home-V4-Cli/home-v4-cli/Model/ExtendModel.php(94): Ypf\Lib\DatabaseV5->select()
//#3 /usr/local/nginx-1.8.1/html/Home-V4-Cli/home-v4-cli/Service/Tsgz/Map/SituationMapService.php(1467): Model\ExtendModel->getListByWhere(Array, 0, 1, 'last_time desc')
//#4 /usr/local/nginx-1.8.1/html/Home-V4-Cli/home-v4-cli/Service/Tsgz/Map/ChinaSituationMapService.php(695): Service\Tsgz\Map\SituationMapService->getNewRiskLog(Array)
//#5 /usr/local/nginx-1.8.1/html/Home-V4-Cli/home-v4-cli/Service/Tsgz/Map/ChinaSituationMapService.php(169): Service\Tsgz\Map\ChinaSituationMapService->getNewT in /usr/local/nginx-1.8.1/html/Home-V4-Cli/Ypf/Lib/DatabaseV5.php on line 98
//日志可以为多行，也可以为单行，完整的一条记录是以 [ 开头
func TailPhpError(queue string, fpath string, timeout int) error {
    isTail := 1
    isWrite := 1
    isRemove :=  0
    finfo , err := os.Stat(fpath)
    if err != nil {
        return err
    }
    seek := finfo.Size()
    tailConfig := tail.Config{Follow: true, MustExist: false, Logger: tail.DiscardingLogger}

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    err = watcher.Add(fpath)
    if err != nil {
        return err
    }
    tailConfig.Location = &tail.SeekInfo{seek, os.SEEK_SET}
    tailHandle, err := tail.TailFile(fpath, tailConfig)
    tmp := ""
    for {
        if isRemove == 1 {
            break
        }
        if isTail == 0 && isWrite == 1 {
            tailConfig.Location = &tail.SeekInfo{seek, os.SEEK_SET}
            tailHandle, _ = tail.TailFile(fpath, tailConfig)
            isTail = 1
        }
        select {
        case line, ok := <- tailHandle.Lines:
            if ok {
                seek = seek + int64(len(line.Text)) + 1
                if strings.HasPrefix(line.Text, "[") {    //独立的一行
                    if tmp != "" {      //文件开始
                        jdata := ParsePhpError(tmp)
                        body, err := json.Encode(jdata)
                        if err == nil {
                            logQueue <- LogRecord{Queue: queue, Body: body}
                        }
                    }
                    tmp = ""
                }
                if tmp != "" {
                    tmp = tmp + "\n"
                }
                tmp = tmp + line.Text
                total++
            }

        case event, ok := <-watcher.Events:
            if ok && event.Op == fsnotify.Write {
                isWrite = 1
            }
            if ok && (event.Op == fsnotify.Remove || event.Op == fsnotify.Rename) {
                isRemove = 1       //删除/重命名
                watcher.Close()
                delete(tailFiles, fpath)
            }

        //case err, ok := <-watcher.Errors:
        //    fmt.Println("goTail watcher error", err, ok)

        case <- time.After(time.Second * time.Duration(timeout)):
            isTail = 0           //超时
            isWrite = 0
            //多行日志
            if tmp != "" {
                jdata := ParsePhpError(tmp)
                body, err := json.Encode(jdata)
                if err == nil {
                    logQueue <- LogRecord{Queue: queue, Body: body}
                }
            }
            tailHandle.Stop()
        }
        if isTail == 0 && isWrite == 0 {
            time.Sleep(time.Duration(1)*time.Second)
        }
    }
    return nil
}

func TailLoges(fpath string, timeout int) error {
    _, err := os.Stat(fpath)
	if os.IsNotExist(err) {
        log.Error(err)
        return err
	}
    npath := fpath + ".tmp"
    os.Rename(fpath, npath)
    fp, err := os.Open(npath)
    if err != nil {
        log.Error(err)
        return err
    }
    reader := bufio.NewReader(fp)
    for {
        line, err := reader.ReadString('\n')
        if err != nil && io.EOF == err {
            break
        }
        row := strings.Split(line, "\t")
        queue := row[1]
        body := []byte(row[2])
        logQueue <- LogRecord{Queue: queue, Body: body}
    }
    delete(tailFiles, fpath)
    os.Remove(npath)
    return nil
}

func DoWatch() {
    //timeout := strings.TrimSpace(viper.GetInt("timeout"))
    timeoutTail := viper.GetInt("tail_timeout")
    docpre := strings.TrimSpace(viper.GetString("docpre"))
    log.Info("start")

    go ToQueue()
    for {
        //遍历监控
        watches := viper.Get("watches").([]interface{})
        for _, tmp := range watches {
            row := tmp.(map[string]interface{})

            //日志目录
            logdir := row["logdir"].(string)
            if !strings.HasSuffix(logdir, "/") {
                logdir = logdir + "/"
            }
            _, err := os.Stat(logdir)
	        if os.IsNotExist(err) {
                log.Error(err)
                continue
	        }

            //匹配日志文件
            glob := row["glob"].(string)
            fpaths, err := filepath.Glob(logdir + glob)
            if err != nil {
                log.Error(logdir + " not exists")
                continue
            }

            ttype := row["type"].(string)
            queue := row["queue"].(string)
            queueTrimPrefix := row["queueTrimPrefix"].(string)
            queueTrimSuffix := row["queueTrimSuffix"].(string)
            for _, fpath := range fpaths {
                //队列名
                fqueue := ""
                fname := path.Base(fpath)
                if queue == "*" {
                    fqueue = strings.TrimPrefix(fname, queueTrimPrefix)
                    fqueue = strings.TrimSuffix(fqueue, queueTrimSuffix)
                } else {
                    fqueue = queue
                }
                cqueue := docpre + fqueue

                //监控
                if _, ok := tailFiles[fpath]; !ok {
                    tailFiles[fpath] = map[string]string{"fpath": fpath, "fname": fname, "queue": fqueue}
                    if ttype == "ypf_error" {
                        go TailYpfError(cqueue, fqueue, fpath, timeoutTail)
                    }
                    if ttype == "php_error" {
                        go TailPhpError(cqueue, fpath, timeoutTail)
                    }
                    if ttype == "log_queue" {
                        go TailLoges(fpath, timeoutTail)
                    }
                }
            }
        }
        time.Sleep(time.Duration(2)*time.Second)
    }
}

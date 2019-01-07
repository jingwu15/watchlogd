package watch

import (
    "os"
    //"log"
    "fmt"
    "time"
    "strings"
    "path/filepath"
    "github.com/spf13/viper"
    "github.com/hpcloud/tail"
    "github.com/fsnotify/fsnotify"
	"github.com/jingwu15/golib/json"
	"github.com/jingwu15/golib/beanstalk"
)

type LogRecord struct {
    Syskey string
    Body  []byte
}

var procExit  = 0
var total = 0
var tailFiles = make(map[string]string)
var logQueue = make(chan LogRecord)
var logpre = "log_phperror_"
var docpre = ""

func ToQueue() {
    for {
        select{
        case log, ok := <- logQueue:
            if ok {
                _, err := beanstalk.Put(docpre + logpre + log.Syskey, log.Body)
                if err != nil {
                    fmt.Println("ToQueue error: ", err)
                }
                total--
            }
        case <- time.After(time.Second * 10):
            //fmt.Println("ToQueue timeout")
        }
    }
}

func TailToQueue(fkey string, fname string) error {
    isTail := 1
    isWrite := 1
    isRemove :=  0
    record := map[string]interface{}{}
    tailConfig := tail.Config{Follow: true, MustExist: false, Location: &tail.SeekInfo{0, os.SEEK_END}, Logger:tail.DiscardingLogger}

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    err = watcher.Add(fname)
    if err != nil {
        return err
    }
    tailHandle, err := tail.TailFile(fname, tailConfig)
    for {
        if isRemove == 1 {
            break
        }
        if isTail == 0 && isWrite == 1 {
            tailHandle, _ = tail.TailFile(fname, tailConfig)
            isTail = 1
        }
        select {
        case line, ok := <- tailHandle.Lines:
            if ok {
                row := strings.Split(line.Text, "\t")
                //json
                err = json.Decode([]byte(row[1]), &record)
                if err == nil {
                    record["syskey"] = fkey
                    body, err := json.Encode(record)
                    if err == nil {
                        logQueue <- LogRecord{Syskey:fkey, Body:body}
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
                delete(tailFiles, fkey)
            }

        //case err, ok := <-watcher.Errors:
        //    fmt.Println("goTail watcher error", err, ok)

        case <- time.After(time.Second * 10):
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

func DoWatch() {
    docpre = strings.TrimSpace(viper.GetString("docpre"))
    logpreTmp := strings.TrimSpace(viper.GetString("logpre"))
    logdir := strings.TrimSpace(viper.GetString("logdir"))
    _, err := os.Stat(logdir)
	if os.IsNotExist(err) {
        fmt.Println(logdir, "not exists")
        os.Exit(1)
	}
    if !strings.HasSuffix(logdir, "/") {
        logdir = logdir + "/"
    }
    if logpreTmp != "" {
        logpre = logpreTmp
    }

    go ToQueue()
    globPrefix := logdir + logpre
    for {
        fnames, err := filepath.Glob(globPrefix + "*.log")
        if err == nil {
            for _, fname := range fnames {
                fkey := strings.TrimPrefix(fname, globPrefix)
                fkey = strings.TrimSuffix(fkey, ".log")
                if _, ok := tailFiles[fkey]; !ok {
                    tailFiles[fkey] = fname
                    go TailToQueue(fkey, fname)
                }
            }
        }
        time.Sleep(time.Duration(10)*time.Second)
    }
}

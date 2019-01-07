### App调度 Cli

```
watchlogd 用户监控日志文件变更，并将变更写入到队列
```

#### 1. 安装
```
# 1. 安装信赖包
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/Shopify/sarama
go get github.com/buger/jsonparser
go get github.com/garyburd/redigo/redis
go get github.com/json-iterator/go
go get github.com/sirupsen/logrus
go get github.com/erikdubbelboer/gspt
go get github.com/oschwald/maxminddb-golang
go get github.com/jinzhu/gorm
go get github.com/beanstalkd/go-beanstalk
go get github.com/jingwu15/golib
或
sh ./install.sh

# 2. 构建项目
go build -o ./bin/watchlogd github.com/jingwu15/watchlogd/watchlogd
```

#### 2. 使用
```
[jingwu@master crontab_sched]$ ./bin/watchlogd

```


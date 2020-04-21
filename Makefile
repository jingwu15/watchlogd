default: build

build:
	go build -o ./bin/watchlogd github.com/jingwu15/watchlogd/cli

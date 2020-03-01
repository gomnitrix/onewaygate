package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var file *os.File
var ELog *log.Logger
var ChMap = make(map[string]chan bool)

func init() {
	file, err := os.OpenFile(logPath+"log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open log failed")
		return
	}
	ELog = log.New(io.MultiWriter(file, os.Stderr), "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	return
}

// 日志文件位置
const logPath = "/C/Users/omnitrix/GoLand/src/controller.com/log/"

// 容器基础信息
const (
	DockerVersion = "1.37"
	Image         = "ubuntu:14.04"
)

// 容器单向隔离配置
const (
	DEFAULTUNIXSOCKET = ""
	RootPw            = "19981017"
	AdoreNgPath       = "/home/omnitrix/gradproj/adore-ng-master/adore-ng.tar"
	HiderTarPath      = "/home/omnitrix/gradproj/controller/bin/controller.tar"
	MountIDPath       = "/var/lib/docker/image/aufs/layerdb/mounts/"
	ContLayerPath     = "/var/lib/docker/aufs/mnt/"
	MntDstPath        = "/mnt"
	ExecFileDstPath   = "/home/"
	AvaPath           = ExecFileDstPath + "ava"
	HiderPath         = ExecFileDstPath + "controller"
	KoPath            = ExecFileDstPath + "adore-ng.ko"
	MaxPrePidsNum     = 5
	ManagerPrefix     = "m"
)

const (
	Addr        = ":8080"
	Host        = "http://127.0.0.1"
	PidFilePath = "/var/run/owmirror.pid"
	AppName     = "owmirror"
	EnvName     = "owm_Daemon"
)

const (
	SleepTime = time.Duration(20) * time.Second
)

const Title = `
  ______          ____  __ _____ _____  _____   ____  _____  
 / __ \ \        / /  \/  |_   _|  __ \|  __ \ / __ \|  __ \ 
| |  | \ \  /\  / /| \  / | | | | |__) | |__) | |  | | |__) |
| |  | |\ \/  \/ / | |\/| | | | |  _  /|  _  /| |  | |  _  / 
| |__| | \  /\  /  | |  | |_| |_| | \ \| | \ \| |__| | | \ \ 
 \____/   \/  \/   |_|  |_|_____|_|  \_\_|  \_\\____/|_|  \_\
`

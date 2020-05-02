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

func init() {
	file, err := os.OpenFile(logPath+"log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open log failed")
		return
	}
	ELog = log.New(io.MultiWriter(file, os.Stderr), "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	return
}

// 日志文件相对位置
const logPath = "/C/Users/omnitrix/GoLand/src/controller.com/log/"
const ServerPath = "api/server/"
const TemplatePath = ServerPath + "template"
const StaticPath = ServerPath + "static"
const CertPath = ServerPath + "certificate/server.crt"
const KeyPath = ServerPath + "certificate/server.key"

// 容器基础信息
const (
	DockerVersion = "1.37"
	Image         = "ubuntu:14.04"
	MgrImage      = "ubuntu:mgr"
)

// 容器单向隔离配置
const (
	DEFAULTUNIXSOCKET = ""
	RootPw            = "19981017"
	AdoreNgPath       = "/home/omnitrix/template/adore-ng-master/adore-ng.tar"
	HiderTarPath      = "/home/omnitrix/template/controller/bin/controller.tar"
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

//server config
const (
	Addr        = ":8080"
	Host        = "http://127.0.0.1"
	PidFilePath = "/var/run/owmirror.pid"
	AppName     = "owmirror"
	EnvName     = "owm_Daemon"
)

//pid isolation interval
const (
	SleepTime = time.Duration(20) * time.Second
)

//iptables config (for network isolation)
const (
	Table     = "mangle"
	PreChain  = "PREROUTING"
	PostChain = "POSTROUTING"
)

// database config
const (
	DbUserName      = "gbw"
	DbPassWd        = "19981017"
	DbNetWork       = "tcp"
	DbServer        = "127.0.0.1"
	DbPort          = 3306
	DbName          = "GRADPROJ"
	ConnMaxLifeTime = time.Duration(200) * time.Second
	MaxOpenConns    = 20
)

const SessionExpires = time.Duration(500) * time.Minute

const Title = `
  ______   ____    __    ____ .___  ___.  __  
 /  __  \  \   \  /  \  /   / |   \/   | |  | 
|  |  |  |  \   \/    \/   /  |  \  /  | |  | 
|  |  |  |   \            /   |  |\/|  | |  | 
|  '--'  |    \    /\    /    |  |  |  | |__| 
 \______/      \__/  \__/     |__|  |__| (__)
`

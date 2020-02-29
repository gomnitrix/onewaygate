package config

import (
	"io"
	"log"
	"os"
	"time"
)

var file *os.File
var ELog *log.Logger
var ChMap map[string]chan bool

func init() {
	file, err := os.OpenFile(logPath+"log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("open log failed")
	}
	ELog = log.New(io.MultiWriter(file, os.Stderr), "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	return
}

const logPath = "/C/Users/omnitrix/GoLand/src/controller.com/log"

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
	DockerVersion     = "1.37"
	Image             = "ubuntu:14.04"
	MaxPrePidsNum     = 5
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

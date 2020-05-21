package daemon //检查pidFile是否存在以及文件里的pid是否存活

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	server2 "controller.com/api/server"

	"controller.com/config"
)

//improvement http.Server
type server struct {
	http.Server
	listener *listener
	cm       *ConnectionManager
}

//用来重载net.Listener的方法
type listener struct {
	net.Listener
	server *server
}

var (
	TimeDeadLine = 10 * time.Second
	srv          *server
	appName      = config.AppName
	pidFile      = config.PidFilePath
	pidVal       int
)

func init() {
	if os.Getenv(config.EnvName) != "true" {
		cmd := ""
		if l := len(os.Args); l > 1 {
			cmd = os.Args[l-1]
			if (cmd == "--daemon" || cmd == "--debug" || cmd == "-d" || cmd == "stop" || cmd == "-h") && l > 2 {
				fmt.Printf("Usage: %s, to use server: -d|daemon|stop|-h\n", appName)
				os.Exit(0)
			}
		}
		switch cmd {
		case "--debug":
			config.Debug = true
			return
		case "--daemon":
			fallthrough
		case "-d":
			if isRunning() {
				log.Printf("[%d] %s is already running\n", pidVal, appName)
			} else { //fork daemon进程
				if err := forkDaemon(); err != nil {
					log.Fatal(err)
				}
			}
		case "stop": //停止
			if !isRunning() {
				log.Printf("%s is not running\n", appName)
			} else {
				syscall.Kill(pidVal, syscall.SIGTERM) //kill
			}
		case "-h":
			fmt.Printf("Usage: %s start|restart|stop\n", appName)
		default: //其它不识别的参数
			return //返回至调用方
		}
		//主进程退出
		os.Exit(0)
	}
	go handleSignals()
}

func isRunning() bool {
	if mf, err := os.Open(pidFile); err == nil {
		pid, _ := ioutil.ReadAll(mf)
		pidVal, _ = strconv.Atoi(string(pid))
	}
	running := false
	if pidVal > 0 {
		if err := syscall.Kill(pidVal, 0); err == nil { //发一个信号为0到指定进程ID，如果没有错误发生，表示进程存活
			running = true
		}
	}
	return running
}

//保存pid
func savePid(pid int) error {
	file, err := os.OpenFile(pidFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(strconv.Itoa(pid))
	return nil
}

//捕获系统信号
func handleSignals() {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	var err error
	for {
		sig := <-signals
		switch sig {
		case syscall.SIGHUP: //重启
			if srv != nil {
				err = srv.fork()
			} else { //only deamon时不支持kill -HUP,因为可能监听地址会占用
				log.Printf("[%d] %s stopped.", os.Getpid(), appName)
				os.Remove(pidFile)
				os.Exit(2)
			}
			if err != nil {
				log.Fatalln(err)
			}
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			fmt.Printf("[%d] %s stop graceful", os.Getpid(), appName)
			if srv != nil {
				srv.shutdown()
			} else {
				fmt.Printf("[%d] %s stopped.", os.Getpid(), appName)
			}
			os.Exit(1)
		}
	}
}

//forkDaemon,当checkPid为true时，检查是否有存活的，有则不执行
func forkDaemon() error {
	args := os.Args
	os.Setenv(config.EnvName, "true")
	procAttr := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}
	pid, err := syscall.ForkExec(os.Args[0], args, procAttr)
	if err != nil {
		return err
	}
	log.Printf("[%d] %s start daemon\n", pid, appName)
	savePid(pid)
	return nil
}

//fork一个新的进程
func (this *server) fork() error {
	os.Setenv("_GRACEFUL_RESTART", "true")
	lFd, err := this.listener.File()
	if err != nil {
		return err
	}
	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), lFd},
	}
	pid, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
	if err != nil {
		return err
	}
	savePid(pid)
	log.Printf("[%d] %s fork ok\n", pid, appName)
	return nil
}

//关闭服务
func (this *server) shutdown() {
	server2.CloseDb()
	this.SetKeepAlivesEnabled(false)
	this.cm.close(TimeDeadLine)
	this.listener.Close()
	fmt.Printf("[%d] %s stopped.", os.Getpid(), appName)
}

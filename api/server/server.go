package server

import (
	"fmt"
	"io"
	"net/http"

	"controller.com/internal/app/netisol"

	"controller.com/config"
	"controller.com/internal"
	"controller.com/internal/app/mntisol"
	"controller.com/internal/app/pidisol"
)

type HttpHanlder func(w http.ResponseWriter, req *http.Request)

var log = config.ELog

type server struct {
	in     io.ReadCloser
	out    io.Writer
	err    io.Writer
	addr   string
	router map[string]HttpHanlder
}

func (s server) StartServe() {
	mux := http.NewServeMux()
	for route, handler := range s.router {
		mux.HandleFunc(route, handler)
	}
	//TODO add log
	if err := http.ListenAndServe(s.addr, mux); err != nil {
		//TODO log
		fmt.Fprintf(s.out, err.Error())
	}
}

func NewServer(addr string, in io.ReadCloser, out, err io.Writer) server {
	ns := server{
		addr:   addr,
		router: nil,
		in:     in,
		out:    out,
		err:    err,
	}
	ns.router = map[string]HttpHanlder{
		"/run":  RunHandler,
		"/stop": StopHandler,
	}
	return ns
}
func RunHandler(w http.ResponseWriter, r *http.Request) {
	param, err := internal.GetParamJson(r)
	if err != nil {
		log.Println(err.Error())
		return
	}
	target := param["target"]
	if target != "" {
		// TODO check the target ID
		fmt.Fprintf(w, "the target:%s\n", target)
		fmt.Fprint(w, "Trying to create new manager for this target...\n")
	} else {
		fmt.Fprint(w, "no target, trying to create\n")
		target = internal.CreateTarget()
	}
	manager := internal.CreateRunManager(target)
	fmt.Fprintf(w, "Create successfully:\nthe manager ID is %s\nthe target ID is %s\n", manager, target)

	// init the isolation relationship between manager and target
	go pidisol.PidIsolation(manager)
	mntisol.MountIsolation(manager, target)
	netisol.NetWorkIsolation(manager, target)
	//TODO add User ns isolation
}

func StopHandler(w http.ResponseWriter, r *http.Request) {
	param, err := internal.GetParamJson(r)
	if err != nil {
		log.Println(err.Error())
		return
	}
	manager := param["manager"]
	pidisol.SendStopToChan(manager)
	mntisol.UmountTarget(manager)

	for _, cont := range param {
		if cont == "" {
			continue
		}
		err := internal.RmContainer(cont)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "remove %s failed\n", cont)
		}
		fmt.Fprintf(w, "%s removed\n", cont)
	}
}

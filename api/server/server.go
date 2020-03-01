package server

import (
	"fmt"
	"io"
	"net/http"

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
		"/run": RunHandler,
	}
	return ns
}
func RunHandler(w http.ResponseWriter, r *http.Request) {
	param := internal.GetParamJson(r)
	target := param["target"]
	if target != "" {
		// TODO check the target ID
		fmt.Fprint(w, "the target:", target)
		fmt.Fprint(w, "Trying to create new manager for the target...")
	} else {
		fmt.Fprint(w, "no target, trying to create.")
		target = internal.CreateTarget()
	}
	manager := internal.CreateRunManager(target)
	fmt.Fprint(w, "Create successfully, the manager ID is ", manager)

	// init the isolation relationship between manager and target
	go pidisol.PidIsolation(manager)
	//mntisol.MountIsolation(manager, target)
	//TODO add User ns and Net ns isolation
}

func StopHandler(w http.ResponseWriter, r *http.Request) {
	param := internal.GetParamJson(r)
	manager := param["manager"]
	pidisol.SendStopToChan(manager)
	mntisol.UmountTarget(manager)

	for _, cont := range param {
		err := internal.RmContainer(cont)
		if err != nil {
			log.Println(err)
		}
	}
}

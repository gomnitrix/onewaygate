package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"controller.com/internal"
	"controller.com/internal/app/mntisol"
	"controller.com/internal/app/pidisol"
)

type HttpHanlder func(w http.ResponseWriter, req *http.Request)

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

func RunHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var param map[string]string
	decoder.Decode(param)
	target := param["target"]
	if target != "" {
		// TODO check the target ID
		fmt.Println("the target:", target)
		fmt.Println("Trying to create new manager for the target...")
	} else {
		fmt.Println("no target, trying to create.")
		target = internal.CreateTarget()
	}
	manager := internal.CreateRunManager(target)
	fmt.Println("Create successfully, the manager ID is ", manager)

	// init the isolation relationship between manager and target
	pidisol.PidIsolation(manager)
	mntisol.MountIsolation(manager, target)
	//TODO add User ns and Net ns isolation
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

package server

import (
	"fmt"
	"io"
	"net/http"

	"controller.com/config"
	"controller.com/internal"
	"controller.com/internal/app/mntisol"
	"controller.com/internal/app/netisol"
	"controller.com/internal/app/pidisol"
	"controller.com/internal/app/sqlhelper"

	"github.com/kataras/iris"
)

type HttpHanlder func(w http.ResponseWriter, req *http.Request)

var log = config.ELog
var myDb *sqlhelper.DbHelper

func initDb() {
	if myDb != nil {
		return
	}
	myDb = sqlhelper.GetNewHelper()
}

type server struct {
	in     io.ReadCloser
	out    io.Writer
	err    io.Writer
	addr   string
	router map[string]HttpHanlder
}

// old func
func (s server) StartServe() {
	mux := http.NewServeMux()
	for route, handler := range s.router {
		mux.HandleFunc(route, handler)
	}
	//TODO add log
	pidisol.InitMap()
	initDb()
	if err := http.ListenAndServe(s.addr, mux); err != nil {
		//TODO log
		fmt.Fprintf(s.out, err.Error())
	}
}

func StartServe(app *iris.Application) {
	pidisol.InitMap()
	initDb()
	app.Post("/run", NewRunHandler)
	app.Post("/stop", NewStopHandler)
	if err := app.Run(iris.Addr(config.Addr)); err != nil {
		//TODO fix log
		fmt.Println(err.Error())
	}
}

//func NewServer(addr string, in io.ReadCloser, out, err io.Writer) server {
//	ns := server{
//		addr:   addr,
//		router: nil,
//		in:     in,
//		out:    out,
//		err:    err,
//	}
//	ns.router = map[string]HttpHanlder{
//		"/run":  RunHandler,
//		"/stop": StopHandler,
//	}
//	return ns
//}

func NewRunHandler(ctx iris.Context) {
	var payload map[string]string
	err := ctx.ReadJSON(&payload)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}
	target := payload["target"]
	targetName := payload["TgtName"]
	if target != "" {
		// TODO check the target ID
		ctx.Writef("the target:%s\n", target)
		ctx.Writef("Trying to create new manager for this target...\n")
	} else {
		ctx.Writef("no target, trying to create\n")
		target = internal.CreateTarget(targetName)
	}
	manager := internal.CreateRunManager(target)
	ctx.Writef("Create successfully:\nthe manager ID is %s\nthe target ID is %s\n", manager, target)

	// init the isolation relationship between manager and target
	myDb.InputConts(target, manager)
	go pidisol.PidIsolation(manager)
	mntisol.MountIsolation(manager, target)
	netisol.NetWorkIsolation(manager, target)
	//TODO add User ns isolation
	ctx.Writef("All isolation done.\n")
}

func NewStopHandler(ctx iris.Context) {
	var payload map[string]string
	err := ctx.ReadJSON(&payload)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}
	manager := payload["manager"]
	target := payload["target"]
	pidisol.SendStopToChan(manager)
	mntisol.UmountTarget(manager)
	netisol.RmTeeRules(manager, target)
	conts := [2]string{manager, target}

	for _, cont := range conts {
		if cont == "" {
			continue
		}
		err := internal.RmContainer(cont)
		if err != nil {
			log.Println(err)
			ctx.Writef("remove %s failed\n", cont)
		}
		ctx.Writef("%s removed\n", cont)
	}
	myDb.DeleteConts(target)
}

//func WebRunHandler(ctx iris.Context) {
//
//}

//func RunHandler(w http.ResponseWriter, r *http.Request) {
//	param, err := internal.GetParamJson(r)
//	if err != nil {
//		log.Println(err.Error())
//		return
//	}
//	target := param["target"]
//	targetName := param["TgtName"]
//	if target != "" {
//		// TODO check the target ID
//		fmt.Fprintf(w, "the target:%s\n", target)
//		fmt.Fprint(w, "Trying to create new manager for this target...\n")
//	} else {
//		fmt.Fprint(w, "no target, trying to create\n")
//		target = internal.CreateTarget(targetName)
//	}
//	manager := internal.CreateRunManager(target)
//	fmt.Fprintf(w, "Create successfully:\nthe manager ID is %s\nthe target ID is %s\n", manager, target)
//
//	// init the isolation relationship between manager and target
//	myDb.InputConts(target, manager)
//	go pidisol.PidIsolation(manager)
//	mntisol.MountIsolation(manager, target)
//	netisol.NetWorkIsolation(manager, target)
//	//TODO add User ns isolation
//}
//
//func StopHandler(w http.ResponseWriter, r *http.Request) {
//	param, err := internal.GetParamJson(r)
//	if err != nil {
//		log.Println(err.Error())
//		return
//	}
//	manager := param["manager"]
//	target := param["target"]
//	pidisol.SendStopToChan(manager)
//	mntisol.UmountTarget(manager)
//	netisol.RmTeeRules(manager, target)
//
//	for _, cont := range param {
//		if cont == "" {
//			continue
//		}
//		err := internal.RmContainer(cont)
//		if err != nil {
//			log.Println(err)
//			fmt.Fprintf(w, "remove %s failed\n", cont)
//		}
//		fmt.Fprintf(w, "%s removed\n", cont)
//	}
//	myDb.DeleteConts(target)
//}

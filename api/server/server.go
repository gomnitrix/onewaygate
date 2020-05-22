package server

import (
	"errors"
	"fmt"

	"github.com/kataras/iris/middleware/logger"

	"controller.com/internal/OwmError"

	"github.com/kataras/iris/sessions"

	"controller.com/config"
	"controller.com/internal"
	"controller.com/internal/app/mntisol"
	"controller.com/internal/app/netisol"
	"controller.com/internal/app/pidisol"
	"controller.com/internal/app/sqlhelper"
	"github.com/kataras/iris"
)

var log = config.ELog
var myDb *sqlhelper.DbHelper
var certPath = internal.JoinPath(config.CertPath)
var keyPath = internal.JoinPath(config.KeyPath)
var logPath = config.LogPath
var debug bool

var (
	cookieNameForSessionID = "owmsessionid"
	sess                   = sessions.New(sessions.Config{
		Cookie:  cookieNameForSessionID,
		Expires: config.SessionExpires,
	})
)

func StartServe(app *iris.Application) {
	debug = config.Debug
	f := newLogFile()
	defer f.Close()
	ChooseLogOutput(app, f)
	requestLogger := logger.New(logger.Config{
		Status:             true,
		IP:                 true,
		Method:             true,
		Path:               true,
		Query:              true,
		MessageContextKeys: []string{"logger_message"},
		MessageHeaderKeys:  []string{"User-Agent"},
	})
	app.UseGlobal(requestLogger)

	pidisol.InitMap()
	initDb()
	if err := app.Run(iris.TLS(config.Addr, certPath, keyPath)); err != nil {
		if debug {
			fmt.Println(err.Error())
		}
		app.Logger().Errorf("server.go: Func StartServe, line 66. Error: %s", err.Error())
	}
}

func NewRunHandler(ctx iris.Context) {
	var payload map[string]string
	err := ctx.ReadJSON(&payload)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}
	//target := payload["target"]
	targetName := payload["TgtName"]
	resp := CreateContGroup("Admin", targetName, "")
	ctx.Writef(resp)
}

func NewRemoveHandler(ctx iris.Context) {
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

func Error404(ctx iris.Context) {
	ErrorHandler(ctx, "404", "Page Not Found !")
}

func Error403(ctx iris.Context) {
	ErrorHandler(ctx, "403", "You should login !")
}
func Error500(ctx iris.Context) {
	ErrorHandler(ctx, "500", "Ops ! Something wrong, try again !")
}

func WebLoginHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	err := ctx.View("login.html")
	OwmError.Check(err, false, "View page error:login")
}

func LoginHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	session := sess.Start(ctx)
	ctx.ContentType("application/json")
	resp := GetSucceedResponse()
	user := BuildUser(ctx)
	passwd := myDb.QueryPasswd(user.Name)
	if !ComparePasswd(passwd, user.Passwd) {
		OwmError.Check(errors.New("PasswdNotEqualError\n"), false, "Your password is wrong\n")
	}
	session.Set("authenticated", true)
	session.Set("user", user.Name)
	ctx.StatusCode(resp.Status)
	ctx.JSON(resp.Resp)
}

func RegisterHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.ContentType("application/json")
	resp := GetSucceedResponse()
	newUser := BuildUser(ctx)
	//TODO 合法性检查
	myDb.InputUser(newUser)
	ctx.StatusCode(resp.Status)
	ctx.JSON(resp.Resp)
}

func LogoutHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.ContentType("application/json")
	session := sess.Start(ctx)
	resp := GetSucceedResponse()
	session.Set("authenticated", false)
	session.Set("user", nil)
	ctx.StatusCode(resp.Status)
	ctx.JSON(resp.Resp)
}

func CheckSession(ctx iris.Context) {
	defer HandlerRecover(ctx)
	if auth, _ := sess.Start(ctx).GetBoolean("authenticated"); !auth {
		OwmError.Check(OwmError.GetAccessDeniedError("You Need To Login Firstly\n"), false, "MiddleWare: CheckSession\n")
	}
	reqUser := ctx.Params().GetDefault("name", "")
	if user := sess.Start(ctx).Get("user"); user != reqUser {
		OwmError.Check(OwmError.GetAccessDeniedError("Access Forbidden\n"), false, "MiddleWare: CheckSession\n")
	}
	ctx.ViewData("userName", reqUser)
	ctx.Next()
}

func WebMainViewHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	//groups := [2](map[string]string){{"manager": "manager1", "1": "target1"}, {"manager": "manager2", "1": "target1", "2": "target2"}}
	groups := GetMainInfoByUser(ctx.Params().Get("name"))
	ctx.ViewData("index", GetMainIndex(len(groups)))
	ctx.ViewData("groupList", internal.GetJsonData(groups))
	err := ctx.View("main.html")
	OwmError.Check(err, false, "View page error: Main")
}

func WebTableViewHandler(ctx iris.Context) {
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	targetList, managerList := GetTableInfosByUser(ctx.Params().Get("name"))
	ctx.ViewData("mgrList", internal.GetJsonData(managerList))
	ctx.ViewData("tgtList", internal.GetJsonData(targetList))
	ctx.ViewData("index", GetTableIndex(len(managerList)))
	if err := ctx.View("table.html"); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
	}
}

func AddHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.ContentType("application/json")
	//target := payload["target"] TODO 从client里删了这个提供已有目标容器的功能
	resp := GetSucceedResponse()
	usrName := ctx.Params().Get("name")
	targetName := ctx.FormValue("target")
	managerName := ctx.FormValue("manager")
	message := CreateContGroup(usrName, targetName, managerName)
	resp.Resp.Message = message
	ctx.StatusCode(resp.Status)
	ctx.JSON(resp.Resp)
}

func WebAddHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	err := ctx.View("add.html")
	OwmError.Check(err, false, "View page error:add")
}

func WebRemoveHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.ContentType("application/json")
	resp := GetSucceedResponse()
	cont := ctx.FormValue("contID")
	contType := ctx.FormValue("contType")
	message := RemoveByType(cont, contType)
	resp.Resp.Message = message
	ctx.StatusCode(resp.Status)
	ctx.JSON(resp.Resp)
}

func WebContainerHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	contId := ctx.URLParamDefault("contId", "")
	contName := ctx.URLParamDefault("contName", "")
	contType := ctx.URLParamDefault("contType", "")
	if contName == "" {
		contName = internal.GetNameByID(contId)
	}
	ctx.ViewData("ContId", contId)
	ctx.ViewData("ContType", contType)
	ctx.ViewData("ContName", contName)
	err := ctx.View("container.html")
	OwmError.Check(err, false, "View page error:container")
}

func WebTerminalHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	contID := ctx.Params().Get("contid")
	ctx.ViewData("contID", contID)
	err := ctx.View("terminal.html")
	OwmError.Check(err, false, "View page error: Terminal")
}

func TerminalHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	contID := ctx.Params().Get("contid")
	AttachTty(ctx, contID)
}

func WebProfileHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.ContentType("text/html")
	err := ctx.View("profile.html")
	OwmError.Check(err, false, "View page error: Profile")
}

func WebInfoHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.ContentType("text/html")
	err := ctx.View("Infos.html")
	OwmError.Check(err, false, "View page error: Info")
}

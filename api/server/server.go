package server

import (
	"encoding/json"
	"errors"
	"fmt"

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

var (
	cookieNameForSessionID = "owmsessionid"
	sess                   = sessions.New(sessions.Config{
		Cookie:  cookieNameForSessionID,
		Expires: config.SessionExpires,
	})
)

func StartServe(app *iris.Application) {
	pidisol.InitMap()
	initDb()
	if err := app.Run(iris.TLS(config.Addr, certPath, keyPath)); err != nil {
		//TODO fix log
		fmt.Println(err.Error())
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

func Error404(ctx iris.Context) {
	ErrorHandler(ctx, "404", "Page Not Found !")
}

func Error403(ctx iris.Context) {
	ErrorHandler(ctx, "403", "You should login !")
}
func Error500(ctx iris.Context) {
	ErrorHandler(ctx, "500", "Ops ! Something wrong, try again !")
}

func CheckSession(ctx iris.Context) {
	defer HandlerRecover(ctx)
	if auth, _ := sess.Start(ctx).GetBoolean("authenticated"); !auth {
		OwmError.Check(OwmError.GetAccessDeniedError("You Need To Login Firstly\n"), false, "MiddleWare: CheckSession\n")
	}
	ctx.Next()
}

func LoginHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	session := sess.Start(ctx)
	ctx.ContentType("application/json")
	resp := GetSucceedResponse()
	user := BuildUser(ctx)
	passwd := myDb.QueryPasswd(user.Name)
	if !ComparePasswd(passwd, user.Passwd) {
		OwmError.Check(errors.New("PasswdNotEqualError"), false, "Your password is wrong")
	}
	session.Set("authenticated", true)
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

func WebRunHandler(ctx iris.Context) {
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	ctx.ViewData("userName", "ifme")
	//groups := map[string]string{"length": "3", "1": "group1", "2": "group2", "3": "group3"}
	ctx.ViewData("index", [2]string{"0", "1"})
	groups := [2](map[string]string){{"manager": "manager1", "1": "target1"}, {"manager": "manager2", "1": "target1", "2": "target2"}}
	bytesGroups, _ := json.Marshal(groups)
	jsonGroups := string(bytesGroups)
	ctx.ViewData("groupList", jsonGroups)
	if err := ctx.View("main.html"); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
	}
}

func WebTableViewHandler(ctx iris.Context) {
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	//forloop := [2]string{"0", "1"}
	//index := [][]string{{"2", "3"}, {"2", "3", "4"}}
	targetList := [2][]map[string]string{
		{
			{
				"ID":     "target111",
				"Name":   "target1",
				"Status": "Normal",
			},
			{
				"ID":     "target222",
				"Name":   "target2",
				"Status": "Normal",
			},
		},
		{
			{
				"ID":     "target333",
				"Name":   "target3",
				"Status": "Normal",
			},
			{
				"ID":     "target444",
				"Name":   "target4",
				"Status": "Normal",
			},
			{
				"ID":     "target555",
				"Name":   "target5",
				"Status": "Normal",
			},
		},
	}
	managerList := [2](map[string]string){
		{
			"ID":     "AAAAA",
			"Name":   "manager1",
			"Status": "Normal",
		},
		{
			"ID":     "BBBBB",
			"Name":   "manager2",
			"Status": "Normal",
		},
	}
	ctx.ViewData("index", [2]int{0, 1})
	ctx.ViewData("mgrList", internal.GetJsonSata(managerList))
	ctx.ViewData("tgtList", internal.GetJsonSata(targetList))
	if err := ctx.View("table.html"); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
	}
}

func WebLoginHandler(ctx iris.Context) {
	defer HandlerRecover(ctx)
	ctx.Gzip(true)
	ctx.ContentType("text/html")
	err := ctx.View("login.html")
	OwmError.Check(err, false, "View page error:login")
}

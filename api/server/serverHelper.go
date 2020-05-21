package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"controller.com/internal"
	"controller.com/internal/app/mntisol"
	"controller.com/internal/app/netisol"
	"controller.com/internal/app/pidisol"

	"github.com/pkg/errors"

	"controller.com/internal/app/sqlhelper"

	"controller.com/Model"
	"controller.com/internal/OwmError"
	"github.com/gorilla/websocket"

	"github.com/kataras/iris"
)

func Reply(ctx iris.Context, resp ResponseWithStatus) {
	ctx.StatusCode(resp.Status)
	ctx.JSON(resp.Resp)
}

func getResponseWithStatus(status int, succeed bool, message string) *ResponseWithStatus {
	return &ResponseWithStatus{
		Status: status,
		Resp: &Response{
			IfSucceed: succeed,
			Message:   message,
		},
	}
}

func GetSucceedResponse() *ResponseWithStatus {
	return getResponseWithStatus(iris.StatusOK, true, "Operation succeed")
}

func GetFailedResponse(status int, message string) *ResponseWithStatus {
	return getResponseWithStatus(status, false, message)
}

//  从上下文中读取表单中数据，并组建为一个User对象返回
func BuildUser(ctx iris.Context) Model.User {
	defer OwmError.Pack()
	newUser := Model.User{}
	err := ctx.ReadForm(&newUser)
	OwmError.Check(err, false, "Load form into User struct failed")
	return newUser
}

func HandlerRecover(ctx iris.Context) {
	if p := recover(); p != nil {
		var flog = ctx.Application().Logger()
		var errCode = iris.StatusInternalServerError
		var errMessage = "some thing wrong,please try again\n"
		if e, ok := p.(OwmError.Error); ok {
			fmt.Printf("%+v", e.Prev)
			flog.Errorf("%+v", e.Prev)
			cause := errors.Cause(e.Prev)
			errMessage = cause.Error()
			switch cause.(type) {
			case OwmError.AccessDeniedError:
				//errCode = iris.StatusForbidden
				ErrorHandler(ctx, "403", errMessage)
				return
			case OwmError.UserExistError:
				errCode = iris.StatusBadRequest
			case OwmError.UserNotExistError:
				//errCode = iris.StatusBadRequest
				errCode = iris.StatusOK
			default:
				//errMessage = "some thing wrong,please try again\n"
				//如果出错的请求是html类的，则通过页面形式返回500错误
				contentType := ctx.GetContentType()
				if contentType == "text/html; charset=UTF-8" {
					ErrorHandler(ctx, "500", errMessage)
					return
				}
			}
		} else {
			fmt.Sprint(p)
			flog.Error(p)
		}
		failResp := GetFailedResponse(errCode, errMessage)
		Reply(ctx, *failResp)
	}
}

func ComparePasswd(passInDb string, passInReq string) bool {
	if strings.Compare(passInDb, passInReq) != 0 {
		return false
	}
	return true
}

func initDb() {
	if myDb != nil {
		return
	}
	myDb = sqlhelper.GetNewHelper()
}

func CloseDb() {
	myDb.Close()
}

// 按天生成日志文件
func todayFilename() string {
	today := time.Now().Format("20060102")
	return today + ".log"
}

// 创建打开文件
func newLogFile() *os.File {

	filename := internal.JoinPath(logPath + todayFilename())
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	return f
}

func ErrorHandler(ctx iris.Context, errCode string, errMessage string) {
	defer HandlerRecover(ctx)
	ctx.ViewData("message", errMessage)
	err := ctx.View(errCode + ".html")
	OwmError.Check(err, false, "return %s page failed", errCode)
}

func CreateContGroup(userName, tgtName, mgrName string) (resp string) {
	defer OwmError.Pack()
	target := internal.CreateTarget(tgtName)
	manager := internal.CreateRunManager(target, mgrName)
	resp = fmt.Sprintf("Create successfully:\nthe manager ID is %s\nthe target ID is %s\n", manager, target)

	// init the isolation relationship between manager and target
	myDb.InputConts(userName, target, manager)
	go pidisol.PidIsolation(manager)
	mntisol.MountIsolation(manager, target)
	netisol.NetWorkIsolation(manager, target)
	//TODO add User ns isolation

	return
}

func RemoveContGroup(target, manager string) (resp string) {
	defer OwmError.Pack()
	//TODO 检查参数合法性
	pidisol.SendStopToChan(manager)
	mntisol.UmountTarget(manager)
	netisol.RmTeeRules(manager, target)
	//TODO 改成一对多的话要注意这里（上下都是）
	conts := [2]string{manager, target}
	resp = ""
	for _, cont := range conts {
		err := internal.RmContainer(cont)
		OwmError.Check(err, false, "remove %s failed\n", cont)
		resp = resp + fmt.Sprintf("%s removed\n", cont)
	}
	myDb.DeleteConts(target)
	return
}

func RemoveByType(contID, contType string) string {
	defer OwmError.Pack()
	var filteredTgts, targets []string
	var manager string
	if contType == "manager" {
		targets = myDb.GetTargetsByMgr(contID)
		manager = contID
	} else if contType == "target" {
		manager = myDb.GetManagerByTgt(contID)
		targets = append(targets, contID)
	}

	for _, target := range targets {
		filteredTgts = append(filteredTgts, internal.FilterContainerID(target))
	}
	manager = internal.FilterContainerID(manager)
	message := RemoveContGroup(filteredTgts[0], manager) //TODO 如果改成一对多，这里要改成把所有targets都传入
	return message
}

func GetMainInfoByUser(userName string) []map[string]string {
	defer OwmError.Pack()
	var contsInfo [](map[string]string)
	managers := myDb.GetMgrsByUser(userName)
	for _, manager := range managers {
		var tmpInfo = make(map[string]string)
		tmpInfo["manager"] = manager
		tmpTgtsID := myDb.GetTargetsByMgr(manager)
		for idx, target := range tmpTgtsID {
			tmpInfo[strconv.Itoa(idx+1)] = target
		}
		contsInfo = append(contsInfo, tmpInfo)
	}
	return contsInfo
}

func GetTableInfosByUser(userName string) ([][]map[string]string, [](map[string]string)) {
	defer OwmError.Pack()
	var mgrsInfo [](map[string]string)
	var tgtsInfo [][]map[string]string
	var tmpTgtsID []string
	managers := myDb.GetMgrsByUser(userName)
	var cont = Model.ContainerRow{}
	for _, manager := range managers {
		var tmpTgtsInfo [](map[string]string)
		internal.GetContTableInfo(manager, &cont)
		mgrsInfo = append(mgrsInfo, internal.StoM(cont))
		tmpTgtsID = myDb.GetTargetsByMgr(cont.ID)

		for _, target := range tmpTgtsID {
			internal.GetContTableInfo(target, &cont)
			tmpTgtsInfo = append(tmpTgtsInfo, internal.StoM(cont))
		}
		tgtsInfo = append(tgtsInfo, tmpTgtsInfo)
	}
	return tgtsInfo, mgrsInfo
}
func GetMainIndex(num int) []string {
	var tmp []string
	for i := 0; i < num; i++ {
		tmp = append(tmp, strconv.Itoa(i))
	}
	return tmp
}

func GetTableIndex(num int) []int {
	var tmp []int
	for i := 0; i < num; i++ {
		tmp = append(tmp, i)
	}
	return tmp
}

func WriterCopy(reader io.Reader, writer *websocket.Conn) {
	defer OwmError.Pack()
	buf := make([]byte, 8192)
	for {
		nr, err := reader.Read(buf)
		if err != nil {
			return
		}
		//OwmError.Check(err, false, "Read failed\n")
		if nr > 0 {
			err = writer.WriteMessage(websocket.BinaryMessage, buf[0:nr])
			if err != nil {
				return
			}
			//OwmError.Check(err, false, "WriteMessage failed\n")
		}
	}
}

func ReaderCopy(reader *websocket.Conn, writer io.Writer) {
	for {
		messageType, p, err := reader.ReadMessage()
		if err != nil {
			return
		}
		//OwmError.Check(err, false, "Read failed\n")
		if messageType == websocket.TextMessage {
			writer.Write(p)
		}
	}
}

func AttachTty(ctx iris.Context, contID string) {
	defer OwmError.Pack()
	tty := internal.GetTty(contID)
	w := ctx.ResponseWriter()
	r := ctx.Request()

	mytmpu := new(websocket.Upgrader)
	mytmpu.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := mytmpu.Upgrade(w, r, nil)
	OwmError.Check(err, false, "Upgrade failed\n")
	defer conn.Close()
	// 关闭I/O流
	defer tty.Close()
	// 退出进程
	defer func() {
		tty.Conn.Write([]byte("exit\r"))
	}()

	// 转发输入/输出至websocket
	go func() {
		WriterCopy(tty.Conn, conn)
	}()
	ReaderCopy(conn, tty.Conn)
}

package server

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"controller.com/internal/app/sqlhelper"

	"controller.com/Model"
	"controller.com/internal/OwmError"
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
		var errCode = iris.StatusInternalServerError
		var errMessage = "some thing wrong,please try again\n"
		if e, ok := p.(OwmError.Error); ok {
			fmt.Printf("%+v", e.Prev)
			//TODO log
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
				errMessage = "some thing wrong,please try again\n"
				//如果出错的请求是html类的，则通过页面形式返回500错误
				contentType := ctx.GetContentType()
				if contentType == "text/html; charset=UTF-8" {
					ErrorHandler(ctx, "500", errMessage)
					return
				}
			}
		} else {
			//TODO log normal error
			fmt.Sprint(p)
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

func ErrorHandler(ctx iris.Context, errCode string, errMessage string) {
	defer HandlerRecover(ctx)
	ctx.ViewData("message", errMessage)
	err := ctx.View(errCode + ".html")
	OwmError.Check(err, false, "return %s page failed", errCode)
}

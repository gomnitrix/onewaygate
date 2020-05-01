package server

import (
	"fmt"

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

func getFailedResponse(status int, message string) *ResponseWithStatus {
	return getResponseWithStatus(status, false, message)
}

func BuildUser(ctx iris.Context) Model.User {
	defer OwmError.Pack()
	newUser := Model.User{}
	err := ctx.ReadForm(&newUser)
	OwmError.Check(err, false, "Load register form into User struct failed")
	return newUser
}

func HandlerRecover(ctx iris.Context) {
	if p := recover(); p != nil {
		//TODO log
		fmt.Printf("%+v\n", p)
		failResp := getFailedResponse(iris.StatusInternalServerError, fmt.Sprint(p))
		Reply(ctx, *failResp)
	}
}

package Model

type User struct {
	Name   string `form:"usrname"`
	Passwd string `form:"passwd"`
}

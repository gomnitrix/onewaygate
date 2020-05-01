package server

type Response struct {
	IfSucceed bool   `json:"ifSucceed"`
	Message   string `json:"message"`
}

type ResponseWithStatus struct {
	Status int
	Resp   *Response
}

package response

import (
	"encoding/json"
	"fmt"
)

type Response struct {
	code int         `json:"code"` //在json转换中的字段名
	msg  string      `json:"msg"`
	data interface{} `json:"data"`
}

func NewRespone(code int, msg string, data interface{}) Response {
	return Response{
		code: code,
		msg:  msg,
		data: data,
	}
}

func (r Response) ToJson() []byte {
	rjosn, err := json.Marshal(r)
	if err != nil {
		fmt.Println("tojson err:", err)
		return nil
	}
	return rjosn
}

func (r Response) ToJsonString() string {
	rjosn, err := json.Marshal(r)
	if err != nil {
		fmt.Println("tojson err:", err)
		return ""
	}
	return string(rjosn)
}

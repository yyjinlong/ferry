// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package base

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

const (
	Success int = 0
	Failed  int = 1
)

var (
	Msg  string      = ""
	Data interface{} = nil
)

type Executer interface {
	Handle(c *gin.Context, r *MyRequest) (interface{}, error)
}

func Construct(e Executer, c *gin.Context) {
	r := &MyRequest{
		URL:       c.Request.URL.Path,
		Method:    c.Request.Method,
		RequestID: uuid.NewV4().String(),
	}

	// 设置响应头
	c.Header("x-requestid", r.RequestID)

	if data, err := e.Handle(c, r); err != nil {
		Response(c, Failed, err.Error(), Data)
	} else {
		Response(c, Success, Msg, data)
	}
}

func Response(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, MyResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

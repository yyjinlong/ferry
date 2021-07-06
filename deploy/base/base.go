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

type ERRCode struct {
	Success int
	Failed  int
}

type ERRInfo struct {
	Msg  string
	Data interface{}
}

type Executer interface {
	Handle(c *gin.Context, r *MyRequest) (interface{}, error)
}

func Construct(e Executer, c *gin.Context) {
	errCode := ERRCode{0, 1}
	errInfo := ERRInfo{"success", nil}

	r := &MyRequest{
		URL:       c.Request.URL.Path,
		Method:    c.Request.Method,
		RequestID: uuid.NewV4().String(),
	}

	// 设置响应头
	c.Header("x-requestid", r.RequestID)

	data, err := e.Handle(c, r)
	if err != nil {
		response(c, errCode.Failed, err.Error(), errInfo.Data)
		return
	}
	response(c, errCode.Success, errInfo.Msg, data)
}

func response(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, MyResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

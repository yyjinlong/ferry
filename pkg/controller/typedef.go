// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	Success int = 0
	Failed  int = 1
)

// MyResponse http接口响应体
type MyResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Response 响应信息
func Response(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, MyResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

// ResponseSuccess 简化响应信息
func ResponseSuccess(c *gin.Context, data interface{}) {
	Response(c, Success, "success", data)
}

// ResponseFailed 简化响应信息
func ResponseFailed(c *gin.Context, errmsg string) {
	Response(c, Failed, errmsg, nil)
}

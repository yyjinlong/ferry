// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package api

import (
	"html"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

const (
	Success int = 0
	Failed  int = 1
)

type Request struct {
	Ctx     *gin.Context
	TraceID string
}

// Escape 避免xss注入
func (r *Request) Escape(val string) string {
	return html.EscapeString(val)
}

func (r *Request) GetEscape(key string) string {
	return r.Escape(r.Ctx.Query(key))
}

func (r *Request) PostEscape(key string) string {
	return r.Escape(r.Ctx.PostForm(key))
}

// ShouldBind 请求参数为空校验
func (r *Request) ShouldBind(obj interface{}) error {
	return r.Ctx.ShouldBind(obj)
}

func (r *Request) ShouldBindJSON(obj interface{}) error {
	return r.Ctx.ShouldBindJSON(obj)
}

func (r *Request) BindJSON(obj interface{}) error {
	return r.Ctx.BindJSON(obj)
}

// Response 响应信息
func (r *Request) Response(code int, msg string, data interface{}) {
	Response(r.Ctx, code, msg, data)
}

// ResponseSuccess 简化响应信息
func (r *Request) ResponseSuccess(data interface{}) {
	Response(r.Ctx, Success, "success", data)
}

func (r *Request) String(code int, format string, values ...interface{}) {
	r.Ctx.String(code, format, values)
}

type MyResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func UniqueID() string {
	return uuid.NewV4().String()
}

func Response(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, MyResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

type ExtendFunc func(*Request)

func ExtendContext(fn ExtendFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := UniqueID()

		c.Header("x-traceid", traceID)
		r := &Request{c, traceID}

		fn(r)
	}
}

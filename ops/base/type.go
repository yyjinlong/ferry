// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package base

type MyRequest struct {
	Method      string
	URL         string
	RequestID   string
	RequestType int
}

type MyResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

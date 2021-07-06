// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"ferry/deploy/base"
)

type Build struct {
}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	fmt.Println("service build....")
	return "", nil
}

type Query struct {
}

func (q *Query) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	fmt.Println("service query....")
	return "", nil
}

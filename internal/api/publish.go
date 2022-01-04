// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package api

import (
	"github.com/gin-gonic/gin"

	"ferry/internal/bll/publish"
	"ferry/pkg/base"
)

func BuildTag(c *gin.Context) {
	base.Construct(&publish.BuildTag{}, c)
}

func BuildImage(c *gin.Context) {
	base.Construct(&publish.BuildImage{}, c)
}

func Service(c *gin.Context) {
	base.Construct(&publish.Service{}, c)
}

func Deploy(c *gin.Context) {
	base.Construct(&publish.Deploy{}, c)
}

func Finish(c *gin.Context) {
	base.Construct(&publish.Finish{}, c)
}
// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package api

import (
	"github.com/gin-gonic/gin"

	"ferry/server/base"
	"ferry/server/blls/service"
)

func ServiceBuild(c *gin.Context) {
	base.Construct(&service.Build{}, c)
}

func ServiceQuery(c *gin.Context) {
	base.Construct(&service.Query{}, c)
}

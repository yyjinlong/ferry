// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package v1

import (
	"github.com/gin-gonic/gin"

	"ferry/web/base"
	"ferry/web/blls/service"
)

func ServiceBuild(c *gin.Context) {
	base.Construct(&service.Build{}, c)
}

func ServiceQuery(c *gin.Context) {
	base.Construct(&service.Query{}, c)
}

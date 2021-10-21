// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package api

import (
	"github.com/gin-gonic/gin"

	"ferry/server/base"
	"ferry/server/blls/deployment"
)

func DeploymentBuild(c *gin.Context) {
	base.Construct(&deployment.Build{}, c)
}

func DeploymentQuery(c *gin.Context) {
	base.Construct(&deployment.Query{}, c)
}

func DeploymentFinish(c *gin.Context) {
	base.Construct(&deployment.Finish{}, c)
}

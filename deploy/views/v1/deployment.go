// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package v1

import (
	"github.com/gin-gonic/gin"

	"ferry/deploy/base"
	"ferry/deploy/blls/deployment"
)

func DeploymentBuild(c *gin.Context) {
	base.Construct(&deployment.Build{}, c)
}

func DeploymentQuery(c *gin.Context) {
	base.Construct(&deployment.Query{}, c)
}

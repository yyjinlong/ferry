// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package server

import (
	"github.com/gin-gonic/gin"

	"ferry/server/api"
)

func URLs(r *gin.Engine) {
	/* web ui */
	r.GET("/", Index)

	/* api v1 */
	g1 := r.Group("v1")
	{
		g1.POST("/pipeline", api.PipelineCreate)

		g1.GET("/deployment", api.DeploymentQuery)
		g1.POST("/deployment", api.DeploymentBuild)

		g1.POST("/tag", api.BuildTag)
		g1.POST("/image", api.BuildImage)
		g1.POST("/finish", api.DeploymentFinish)

		g1.GET("/service", api.ServiceQuery)
		g1.POST("/service", api.ServiceBuild)
	}
}

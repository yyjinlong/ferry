// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"github.com/gin-gonic/gin"

	"ferry/web/views"
	"ferry/web/views/v1"
)

func urls(r *gin.Engine) {
	/* web ui */
	r.GET("/", views.Index)

	/* api v1 */
	g1 := r.Group("v1")
	{
		g1.POST("/pipeline", v1.PipelineCreate)

		g1.GET("/deployment", v1.DeploymentQuery)
		g1.POST("/deployment", v1.DeploymentBuild)

		g1.POST("/tag", v1.BuildTag)
		g1.POST("/image", v1.BuildImage)
		g1.POST("/finish", v1.DeploymentFinish)

		g1.GET("/service", v1.ServiceQuery)
		g1.POST("/service", v1.ServiceBuild)
	}
}

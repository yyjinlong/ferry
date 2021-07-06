// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"github.com/gin-gonic/gin"

	"ferry/deploy/views"
	"ferry/deploy/views/v1"
)

func urls(r *gin.Engine) {
	/* web ui */
	r.GET("/", views.Index)

	/* api v1 */
	g1 := r.Group("v1")
	{
		g1.GET("/deployment", v1.DeploymentQuery)
		g1.POST("/deployment", v1.DeploymentBuild)

		g1.GET("/service", v1.ServiceQuery)
		g1.POST("/service", v1.ServiceBuild)
	}
}

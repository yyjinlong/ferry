// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nautilus/internal/api"
)

func URLs(r *gin.Engine) {
	/* web ui */
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "show front ui.")

	})

	/* api v1 */
	g1 := r.Group("v1")
	{
		g1.POST("/pipeline", api.CreatePipeline)

		// 发布流程
		g1.POST("/tag", api.BuildTag)
		g1.GET("/tag", api.ReceiveTag)
		g1.POST("/image", api.BuildImage)
		g1.POST("/service", api.Service)
		g1.POST("/deploy", api.Deploy)
		g1.POST("/finish", api.Finish)

	}
}

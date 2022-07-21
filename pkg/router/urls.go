// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nautilus/pkg/controller"
)

func URLs(r *gin.Engine) {
	/* web ui */
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "show front ui.")

	})

	// 上线单
	p := r.Group("v1", UserAuth)
	{
		p.POST("/pipeline", controller.CreatePipeline)
	}

	// 上线流程
	d := r.Group("v1", UserAuth)
	{
		// 发布流程
		d.POST("/tag", controller.BuildTag)
		d.GET("/tag", controller.ReceiveTag)
		d.POST("/image", controller.BuildImage)
		d.POST("/configmap", controller.ConfigMap)
		d.POST("/service", controller.Service)
		d.POST("/deploy", controller.Deploy)
		d.POST("/finish", controller.Finish)
	}

	// 回滚流程
	ro := r.Group("v1", UserAuth)
	{
		ro.POST("/check/rollback", controller.CheckRollback)
		ro.POST("/rollback", controller.Rollback)
	}

	// 定时任务
	cr := r.Group("v1", UserAuth)
	{
		cr.POST("/cronjob", controller.BuildCronjob)
	}
}

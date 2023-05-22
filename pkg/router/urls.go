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
	pipeline := r.Group("v1/pipeline", UserAuth)
	{
		pipeline.POST("/create", controller.CreatePipeline)
	}

	// 上线流程
	deploy := r.Group("v1/deploy", UserAuth)
	{
		// 发布流程
		deploy.POST("/tag", controller.BuildTag)
		deploy.GET("/tag", controller.ReceiveTag)
		deploy.POST("/image/create", controller.BuildImage)
		deploy.GET("/image/update", controller.UpdateImage)
		deploy.POST("/configmap", controller.ConfigMap)
		deploy.POST("/service", controller.Service)
		deploy.POST("/do", controller.Deploy)
		deploy.POST("/finish", controller.Finish)
	}

	// 回滚流程
	rollback := r.Group("v1/rollback", UserAuth)
	{
		rollback.POST("/check", controller.CheckRollback)
		rollback.POST("/do", controller.Rollback)
	}

	// 定时任务
	cron := r.Group("v1/cronjob", UserAuth)
	{
		cron.POST("/create", controller.BuildCronJob)
		cron.POST("/delete", controller.DeleteCronJob)
	}
}

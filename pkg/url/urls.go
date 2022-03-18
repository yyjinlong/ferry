// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package url

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yyjinlong/golib/api"

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
		p.POST("/pipeline", api.ExtendContext(controller.CreatePipeline))
	}

	// 上线流程
	d := r.Group("v1", UserAuth)
	{
		// 发布流程
		d.POST("/tag", api.ExtendContext(controller.BuildTag))
		d.GET("/tag", api.ExtendContext(controller.ReceiveTag))
		d.POST("/image", api.ExtendContext(controller.BuildImage))
		d.POST("/service", api.ExtendContext(controller.Service))
		d.POST("/deploy", api.ExtendContext(controller.Deploy))
		d.POST("/finish", api.ExtendContext(controller.Finish))
	}
}

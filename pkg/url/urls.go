// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package url

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yyjinlong/golib/api"

	"nautilus/pkg/view"
)

func URLs(r *gin.Engine) {
	/* web ui */
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "show front ui.")

	})

	// 上线单
	pu := r.Group("v1", UserAuth)
	{
		pu.POST("/pipeline", api.ExtendContext(view.CreatePipeline))
	}

	// 上线流程
	du := r.Group("v1", UserAuth)
	{
		// 发布流程
		du.POST("/tag", api.ExtendContext(view.BuildTag))
		du.GET("/tag", api.ExtendContext(view.ReceiveTag))
		du.POST("/image", api.ExtendContext(view.BuildImage))
		du.POST("/service", api.ExtendContext(view.Service))
		du.POST("/deploy", api.ExtendContext(view.Deploy))
		du.POST("/finish", api.ExtendContext(view.Finish))
	}
}

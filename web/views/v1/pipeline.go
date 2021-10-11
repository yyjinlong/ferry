// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package v1

import (
	"github.com/gin-gonic/gin"

	"ferry/web/base"
	"ferry/web/blls/pipeline"
)

func PipelineCreate(c *gin.Context) {
	base.Construct(&pipeline.Build{}, c)
}

func BuildTag(c *gin.Context) {
	base.Construct(&pipeline.BuildTag{}, c)
}

func BuildImage(c *gin.Context) {
	base.Construct(&pipeline.BuildImage{}, c)
}

// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package api

import (
	"github.com/gin-gonic/gin"

	"ferry/server/base"
	"ferry/server/blls/pipeline"
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

// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package v1

import (
	"github.com/gin-gonic/gin"

	"ferry/apps/deploy/blls/pipeline"
	"ferry/ops/base"
)

func PipelineCreate(c *gin.Context) {
	base.Construct(&pipeline.Build{}, c)
}

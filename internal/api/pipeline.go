// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package api

import (
	"github.com/gin-gonic/gin"

	"nautilus/internal/bll/pipeline"
	"nautilus/pkg/base"
)

func CreatePipeline(c *gin.Context) {
	base.Construct(&pipeline.Build{}, c)
}

func ListPipeline(c *gin.Context) {

}

func QueryPipeline(c *gin.Context) {

}

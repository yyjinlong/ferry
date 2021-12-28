// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"github.com/gin-gonic/gin"

	"ferry/pkg/base"
)

type Deploy struct {
}

func (d *Deploy) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	// TODO: 建立websocket，调用deployment发布

	return nil, nil
}

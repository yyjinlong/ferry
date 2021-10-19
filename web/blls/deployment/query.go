// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"ferry/web/base"
)

type Query struct{}

func (q *Query) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	fmt.Println("query....")
	return "", nil
}

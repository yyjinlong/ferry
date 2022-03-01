// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package url

import (
	"github.com/gin-gonic/gin"
)

// DefaultAuth 默认授权
func DefaultAuth(c *gin.Context) {
	c.Next()
}

// UserAuth 用户授权
func UserAuth(c *gin.Context) {
	c.Next()
}

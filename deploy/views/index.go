// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	c.String(http.StatusOK, "show front ui.")
}

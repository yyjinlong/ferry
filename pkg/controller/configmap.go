// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"nautilus/golib/log"
	"nautilus/pkg/config"
	"nautilus/pkg/service/publish"
)

func ConfigMap(c *gin.Context) {
	type params struct {
		Namespace string `form:"namespace" binding:"required"` // 命名空间
		Service   string `form:"service" binding:"required"`   // 服务
		Pair      string `form:"pair" binding:"required"`      // kv键值对
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	var (
		namespace = data.Namespace
		service   = data.Service
		pair      = data.Pair
		pairInfo  map[string]string
	)

	if err := json.Unmarshal([]byte(pair), &pairInfo); err != nil {
		Response(c, Failed, fmt.Sprintf(config.CM_DECODE_DATA_ERROR, err), nil)
		return
	}

	cm := publish.NewConfigMap()
	if err := cm.Handle(namespace, service, pair, pairInfo); err != nil {
		log.ID(cm.Logid).Errorf("publish configmap failed: %+v", err)
		Response(c, Failed, fmt.Sprintf(config.CM_PUBLISH_FAILED, err), nil)
		return
	}
	ResponseSuccess(c, nil)
}

// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

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
		ResponseFailed(c, err.Error())
		return
	}

	var (
		namespace = data.Namespace
		service   = data.Service
		pair      = data.Pair
		pairInfo  map[string]string
	)

	if err := json.Unmarshal([]byte(pair), &pairInfo); err != nil {
		ResponseFailed(c, fmt.Sprintf(config.CM_DECODE_DATA_ERROR, err))
		return
	}

	cm := publish.NewConfigMap()
	if err := cm.Handle(namespace, service, pair, pairInfo); err != nil {
		log.Errorf("publish configmap failed: %+v", err)
		ResponseFailed(c, fmt.Sprintf(config.CM_PUBLISH_FAILED, err))
		return
	}
	ResponseSuccess(c, nil)
}

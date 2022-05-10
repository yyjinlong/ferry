// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"encoding/json"
	"fmt"

	"nautilus/golib/api"
	"nautilus/golib/log"
	"nautilus/pkg/config"
	"nautilus/pkg/service/publish"
)

func ConfigMap(r *api.Request) {
	type params struct {
		Namespace string `form:"namespace" binding:"required"` // 命名空间
		Service   string `form:"service" binding:"required"`   // 服务
		Pair      string `form:"pair" binding:"required"`      // kv键值对
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}

	var (
		namespace = data.Namespace
		service   = data.Service
		pair      = data.Pair
	)
	log.InitFields(log.Fields{"logid": r.TraceID})

	var pairInfo map[string]string
	if err := json.Unmarshal([]byte(pair), &pairInfo); err != nil {
		r.Response(api.Failed, fmt.Sprintf(config.CM_DECODE_DATA_ERROR, err), nil)
		return
	}

	cm := publish.NewConfigMap()
	if err := cm.Handle(namespace, service, pair, pairInfo); err != nil {
		r.Response(api.Failed, fmt.Sprintf(config.CM_PUBLISH_FAILED, err), nil)
		return
	}
	r.ResponseSuccess(nil)
}

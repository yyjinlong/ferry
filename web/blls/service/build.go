// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package service

import (
	"github.com/gin-gonic/gin"

	"ferry/ops/base"
	"ferry/ops/db"
	"ferry/ops/log"
	"ferry/ops/objects"
)

type Build struct {
	namespace string // 命名空间
	name      string // 服务名
	id        int64  // 服务id
}

func (b *Build) validate(c *gin.Context) error {
	type params struct {
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return err
	}

	b.name = data.Service
	return nil
}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := b.validate(c); err != nil {
		return "", err
	}
	log.InitFields(log.Fields{"logid": r.RequestID, "service": b.name})

	si, err := objects.GetServiceInfo(b.name)
	if err != nil {
		return "", err
	}

	b.namespace = si.Namespace.Name
	b.id = si.Service.ID
	appid := objects.GetAppID(b.name, b.id, db.PHASE_ONLINE)
	log.Infof("fetch service appid: %s", appid)

	yam := yaml{
		serviceName:   b.name,
		serviceID:     b.id,
		appid:         appid,
		exposePort:    si.Service.Port,
		containerPort: si.Service.ContainerPort,
	}
	tpl, err := yam.instance()
	if err != nil {
		return "", err
	}
	log.Infof("fetch service tpl: %s", tpl)

	ss := newServices()
	if !ss.exist(b.namespace, b.name) {
		if err := ss.create(b.namespace, tpl); err != nil {
			log.Errorf("create service failed: %s", err)
			return "", err
		}
		log.Infof("create service success")

	} else {
		if err := ss.update(b.namespace, b.name, tpl); err != nil {
			log.Errorf("update service failed: %s", err)
			return "", err
		}
		log.Infof("update service success")
	}
	return "", nil
}

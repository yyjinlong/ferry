// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package service

import (
	"github.com/gin-gonic/gin"

	"ferry/ops/db"
	"ferry/ops/log"
	"ferry/ops/objects"
	"ferry/web/base"
)

type Build struct{}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return "", err
	}
	serviceName := data.Service
	log.InitFields(log.Fields{"logid": r.RequestID, "service": serviceName})

	si, err := objects.GetServiceInfo(serviceName)
	if err != nil {
		return "", err
	}

	var (
		serviceID = si.Service.ID
		namespace = si.Namespace.Name
	)

	appid := objects.GetAppID(serviceName, serviceID, db.PHASE_ONLINE)
	log.Infof("fetch service appid: %s", appid)

	yam := yaml{
		serviceName:   serviceName,
		serviceID:     serviceID,
		appid:         appid,
		exposePort:    si.Service.Port,
		containerPort: si.Service.ContainerPort,
	}
	tpl, err := yam.instance()
	if err != nil {
		return "", err
	}
	log.Infof("fetch service tpl: %s", tpl)

	if err := b.publish(namespace, serviceName, tpl); err != nil {
		return "", err
	}
	log.Infof("build service success")
	return "", nil
}

func (b *Build) publish(namespace, service, tpl string) error {
	ss := newServices()
	if !ss.exist(namespace, service) {
		return ss.create(namespace, tpl)
	}
	return ss.update(namespace, service, tpl)
}

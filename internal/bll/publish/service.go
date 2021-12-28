// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"github.com/gin-gonic/gin"

	"ferry/internal/k8s"
	"ferry/pkg/base"
)

type Service struct {
	namespace   string
	serviceName string
}

func (s *Service) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return "", err
	}
	s.serviceName = data.Service
	log.InitFields(log.Fields{"logid": r.RequestID, "service": s.serviceName})

	serviceObj, err := objects.GetServiceInfo(s.serviceName)
	if err != nil {
		return "", err
	}
	s.namespace = serviceObj.Namespace.Name

	appid := objects.GetAppID(s.serviceName, serviceObj.Service.ID, model.PHASE_ONLINE)
	log.Infof("[service] fetch service appid: %s", appid)

	yaml := k8s.ServiceYaml{
		serviceName:   serviceName,
		serviceID:     serviceObj.Service.ID,
		appid:         appid,
		exposePort:    serviceObj.Service.Port,
		containerPort: serviceObj.Service.ContainerPort,
	}
	tpl, err := yaml.Instance()
	if err != nil {
		return "", err
	}
	log.Infof("[service] fetch service tpl: %s", tpl)

	if err := s.execute(tpl); err != nil {
		return "", err
	}
	log.Infof("[service] build service success")
	return nil, nil
}

func (s *Service) execute() error {
	ss := k8s.NewServices(s.namespace, s.service)
	if !ss.Exist() {
		return ss.Create(tpl)
	}
	return ss.Update(tpl)
}

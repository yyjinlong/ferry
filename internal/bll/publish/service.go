// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"sync"

	"github.com/gin-gonic/gin"

	"ferry/internal/k8s"
	"ferry/internal/model"
	"ferry/internal/objects"
	"ferry/pkg/base"
	"ferry/pkg/log"
)

type Service struct {
	namespace     string
	serviceName   string
	serviceID     int64
	exposePort    int
	containerPort int
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
	s.serviceID = serviceObj.Service.ID
	s.exposePort = serviceObj.Service.Port
	s.containerPort = serviceObj.Service.ContainerPort

	var wg sync.WaitGroup
	wg.Add(len(model.PHASE_NAME_LIST))

	for _, phase := range model.PHASE_NAME_LIST {
		go func(phase string) {
			defer wg.Done()
			s.worker(phase)
		}(phase)
	}
	wg.Wait()
	return nil, nil
}

func (s *Service) worker(phase string) error {
	appid := objects.GetAppID(s.serviceName, s.serviceID, phase)
	log.Infof("[service] fetch service appid: %s", appid)

	yaml := &k8s.ServiceYaml{
		ServiceName:   s.serviceName,
		ServiceID:     s.serviceID,
		Phase:         phase,
		AppID:         appid,
		ExposePort:    s.exposePort,
		ContainerPort: s.containerPort,
	}
	tpl, err := yaml.Instance()
	if err != nil {
		return err
	}
	log.Infof("[service] fetch service tpl: %s", tpl)

	if err := s.execute(tpl); err != nil {
		return err
	}
	log.Infof("[service] build service success")
	return nil
}

func (s *Service) execute(tpl string) error {
	ss := k8s.NewServices(s.namespace, s.serviceName)
	if !ss.Exist() {
		return ss.Create(tpl)
	}
	return ss.Update(tpl)
}

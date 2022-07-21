// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	"golang.org/x/sync/errgroup"

	"nautilus/golib/log"
	"nautilus/pkg/config"
	"nautilus/pkg/k8s/exec"
	"nautilus/pkg/k8s/yaml"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

func NewService() *Service {
	return &Service{Logid: util.UniqueID()}
}

type Service struct {
	Logid string
}

func (s *Service) Handle(serviceName string) error {
	serviceObj, err := model.GetServiceInfo(serviceName)
	if err != nil {
		return fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	nsObj, err := model.GetNamespace(serviceObj.NamespaceID)
	if err != nil {
		return fmt.Errorf(config.DB_QUERY_NAMESPACE_ERROR, err)
	}

	var (
		namespace     = nsObj.Name
		serviceID     = serviceObj.ID
		port          = serviceObj.Port
		containerPort = serviceObj.ContainerPort
	)

	var eg errgroup.Group
	for _, phase := range model.PHASE_NAME_LIST {
		phase := phase
		eg.Go(func() error {
			return s.worker(namespace, serviceName, serviceID, phase, port, containerPort)
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf(config.SVC_WAIT_ALL_SERVICE_ERROR, err)
	}
	return nil
}

func (s *Service) worker(namespace, serviceName string, serviceID int64, phase string, port, containerPort int) error {
	appid := util.GetAppID(serviceName, serviceID, phase)
	log.ID(s.Logid).Infof("fetch service appid: %s", appid)

	svcYaml := &yaml.ServiceYaml{
		ServiceName:   serviceName,
		ServiceID:     serviceID,
		Phase:         phase,
		AppID:         appid,
		ExposePort:    port,
		ContainerPort: containerPort,
	}
	tpl, err := svcYaml.Instance()
	if err != nil {
		return fmt.Errorf(config.SVC_BUILD_SERVICE_YAML_ERROR, err)
	}
	log.ID(s.Logid).Infof("create service: %s mapping tpl: %s", appid, tpl)

	if err := s.execute(namespace, serviceName, tpl); err != nil {
		return fmt.Errorf(config.SVC_K8S_SERVICE_EXEC_FAILED, err)
	}
	log.ID(s.Logid).Infof("build service success")
	return nil
}

func (s *Service) execute(namespace, serviceName, tpl string) error {
	ss := exec.NewServices(namespace, serviceName)
	if !ss.Exist() {
		return ss.Create(tpl)
	}
	return ss.Update(tpl)
}

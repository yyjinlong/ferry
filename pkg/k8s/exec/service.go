// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package exec

import (
	"fmt"

	"github.com/yyjinlong/golib/curl"
	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/config"
)

func NewServices(namespace, name string) *Services {
	return &Services{
		address:   getAddress(namespace),
		namespace: namespace,
		name:      name,
	}
}

type Services struct {
	address   string
	namespace string
	name      string
}

func (s *Services) Exist() bool {
	var (
		url = fmt.Sprintf(config.Config().K8S.Service, s.address, s.namespace) + "/" + s.name
	)
	body, err := curl.Get(url, nil, 5)
	if err != nil {
		log.Infof("check service: %s is not exists", s.name)
		return false
	}
	if err := response(body); err != nil {
		return false
	}
	log.Infof("check service: %s is exist.", s.name)
	return true
}

func (s *Services) Create(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Service, s.address, s.namespace)
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := curl.Post(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request create service api error: %s", err)
		return err
	}
	return response(body)
}

func (s *Services) Update(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Service, s.address, s.namespace) + "/" + s.name
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := curl.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update service api error: %s", err)
		return err
	}
	return response(body)
}

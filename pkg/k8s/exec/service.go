// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package exec

import (
	"fmt"

	"nautilus/golib/curl"
	"nautilus/golib/log"
	"nautilus/pkg/config"
)

func NewServices(namespace, name string) *Services {
	cluster := getCluster(namespace)
	return &Services{
		cluster:   cluster,
		address:   getAddress(cluster),
		namespace: namespace,
		name:      name,
	}
}

type Services struct {
	cluster   string
	address   string
	namespace string
	name      string
}

func (s *Services) Exist() bool {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Service, s.address, s.namespace) + "/" + s.name
		header = getHeader(s.cluster)
	)
	body, err := curl.Get(url, header, 5)
	if err != nil {
		log.Infof("check service: %s is not exist", s.name)
		return false
	}
	if err := response(body); err != nil {
		return false
	}
	log.Infof("check service: %s is exist", s.name)
	return true
}

func (s *Services) Create(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Service, s.address, s.namespace)
		header = getHeader(s.cluster)
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
		header = getHeader(s.cluster)
	)
	body, err := curl.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update service api error: %s", err)
		return err
	}
	return response(body)
}

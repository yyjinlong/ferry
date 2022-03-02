// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package exec

import (
	"encoding/json"
	"fmt"

	"github.com/yyjinlong/golib/curl"
	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/cfg"
)

func NewServices(namespace, name string) *Services {
	return &Services{
		namespace: namespace,
		name:      name,
	}
}

type Services struct {
	namespace string
	name      string
}

func (s *Services) Exist() bool {
	var (
		url = fmt.Sprintf(cfg.Config().K8S.Service, s.namespace) + "/" + s.name
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
		url    = fmt.Sprintf(cfg.Config().K8S.Service, s.namespace)
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
		url    = fmt.Sprintf(cfg.Config().K8S.Service, s.namespace) + "/" + s.name
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := curl.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update service api error: %s", err)
		return err
	}
	return response(body)
}

func response(body string) error {
	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("k8s response body json decode error: %s", err)
		return err
	}

	status, ok := resp["status"].(string)
	if ok && status == "Failure" {
		err := fmt.Errorf("%s", resp["message"].(string))
		log.Errorf("request k8s api failed: %s", err)
		return err
	}
	return nil
}

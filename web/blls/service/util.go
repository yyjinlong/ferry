// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package service

import (
	"encoding/json"
	"fmt"

	"ferry/ops/g"
	"ferry/ops/log"
)

func newServices() *services {
	return &services{}
}

type services struct{}

func (s *services) exist(namespace, name string) bool {
	var (
		url = fmt.Sprintf(g.Config().K8S.Service, namespace) + "/" + name
	)
	body, err := g.Get(url, nil, nil, 5)
	if err != nil {
		log.Infof("check service: %s is not exists", name)
		return false
	}
	if err := s.response(body); err != nil {
		return false
	}
	log.Infof("check service: %s is exist.", name)
	return true
}

func (s *services) create(namespace, tpl string) error {
	var (
		url    = fmt.Sprintf(g.Config().K8S.Service, namespace)
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := g.Post(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request create service api error: %s", err)
		return err
	}
	return s.response(body)
}

func (s *services) update(namespace, name, tpl string) error {
	var (
		url    = fmt.Sprintf(g.Config().K8S.Service, namespace) + "/" + name
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := g.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update service api error: %s", err)
		return err
	}
	return s.response(body)
}

func (s *services) response(body string) error {
	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("response json decode result error: %s", err)
		return err
	}

	status, ok := resp["status"].(string)
	if ok && status == "Failure" {
		err := fmt.Errorf("%s", resp["message"].(string))
		log.Errorf("request service api failed: %s", err)
		return err
	}
	return nil
}

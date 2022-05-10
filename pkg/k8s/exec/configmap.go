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

func NewConfigMap(namespace, name string) *ConfigMap {
	return &ConfigMap{
		address:   getAddress(namespace),
		namespace: namespace,
		name:      name,
	}
}

type ConfigMap struct {
	address   string
	namespace string
	name      string
}

func (cm *ConfigMap) Exist() bool {
	var (
		url = fmt.Sprintf(config.Config().K8S.ConfigMap, cm.address, cm.namespace) + "/" + cm.name
	)
	body, err := curl.Get(url, nil, 5)
	if err != nil {
		log.Infof("check configmap: %s is not exist", cm.name)
		return false
	}
	if err := response(body); err != nil {
		return false
	}
	log.Infof("check configmap: %s is exist", cm.name)
	return true
}

func (cm *ConfigMap) Create(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.ConfigMap, cm.address, cm.namespace)
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := curl.Post(url, header, []byte(tpl), 5)
	if err != nil {
		log.Infof("request create configmap api error: %s", err)
		return err
	}
	return response(body)
}

func (cm *ConfigMap) Update(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.ConfigMap, cm.address, cm.namespace) + "/" + cm.name
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := curl.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update configmap api error: %s", err)
		return err
	}
	return response(body)
}

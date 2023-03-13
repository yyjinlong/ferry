// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package exec

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/util/curl"
)

func NewConfigMap(namespace, name string) *ConfigMap {
	cluster := getCluster(namespace)
	return &ConfigMap{
		cluster:   cluster,
		address:   getAddress(cluster),
		namespace: namespace,
		name:      name,
	}
}

type ConfigMap struct {
	cluster   string
	address   string
	namespace string
	name      string
}

func (cm *ConfigMap) Exist() bool {
	var (
		url    = fmt.Sprintf(config.Config().K8S.ConfigMap, cm.address, cm.namespace) + "/" + cm.name
		header = getHeader(cm.cluster)
	)
	body, err := curl.Get(url, header, 5)
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
		header = getHeader(cm.cluster)
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
		header = getHeader(cm.cluster)
	)
	body, err := curl.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update configmap api error: %s", err)
		return err
	}
	return response(body)
}

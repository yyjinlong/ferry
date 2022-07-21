// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	"nautilus/golib/log"
	"nautilus/pkg/config"
	"nautilus/pkg/k8s/exec"
	"nautilus/pkg/k8s/yaml"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

func NewConfigMap() *ConfigMap {
	return &ConfigMap{Logid: util.UniqueID()}
}

type ConfigMap struct {
	Logid string
}

func (c *ConfigMap) Handle(namespace, service, pair string, pairInfo map[string]string) error {
	configName := util.GetConfigName(service)

	cmYaml := &yaml.ConfigmapYaml{
		Namespace: namespace,
		Name:      configName,
	}
	tpl, err := cmYaml.Instance(pairInfo)
	if err != nil {
		return fmt.Errorf(config.CM_BUILD_YAML_ERROR, err)
	}

	if err := c.execute(namespace, configName, tpl); err != nil {
		return fmt.Errorf(config.CM_K8S_EXEC_FAILED, err)
	}

	if err := model.UpdateConfigMap(service, pair); err != nil {
		return fmt.Errorf(config.CM_UPDATE_DB_ERROR, err)
	}
	log.ID(c.Logid).Infof("create namespace: %s configmap: %s success", namespace, configName)
	return nil
}

func (c *ConfigMap) execute(namespace, name, tpl string) error {
	cMap := exec.NewConfigMap(namespace, name)
	if !cMap.Exist() {
		return cMap.Create(tpl)
	}
	return cMap.Update(tpl)
}

// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/k8s"
)

func NewConfigMap(namespace, service, pair string, data map[string]string) error {
	configName := k8s.GetConfigmapName(service)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: namespace,
		},
		Data: data,
	}

	resource, err := k8s.New(namespace)
	if err != nil {
		return err
	}
	if err := resource.CreateOrUpdateConfigMap(namespace, configMap); err != nil {
		return fmt.Errorf(config.CM_K8S_EXEC_FAILED, err)
	}
	log.Infof("deploy configmap: %s to k8s success", configName)

	if err := model.UpdateConfigMap(service, pair); err != nil {
		return fmt.Errorf(config.CM_UPDATE_DB_ERROR, err)
	}
	log.Infof("record configmap: %s to db success", configName)
	return nil
}

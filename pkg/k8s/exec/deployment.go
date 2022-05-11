// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package exec

import (
	"encoding/json"
	"fmt"

	"nautilus/golib/curl"
	"nautilus/golib/log"
	"nautilus/pkg/config"
)

func NewDeployments(namespace, deployment string) *Deployments {
	return &Deployments{
		address:    getAddress(namespace),
		namespace:  namespace,
		deployment: deployment,
	}
}

type Deployments struct {
	address    string
	namespace  string
	deployment string
}

func (d *Deployments) Exist() bool {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Deployment, d.address, d.namespace) + "/" + d.deployment
		header = getHeader()
	)
	body, err := curl.Get(url, header, 5)
	if err != nil {
		log.Infof("check deployment: %s is not exist", d.deployment)
		return false
	}
	if err := response(body); err != nil {
		return false
	}
	log.Infof("check deplloyment: %s is exist", d.deployment)
	return true
}

func (d *Deployments) Create(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Deployment, d.address, d.namespace)
		header = getHeader()
	)
	body, err := curl.Post(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request create deployment api error: %s", err)
		return err
	}
	return response(body)
}

func (d *Deployments) Update(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Deployment, d.address, d.namespace) + "/" + d.deployment
		header = getHeader()
	)
	body, err := curl.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update deployment api error: %s", err)
		return err
	}
	return response(body)
}

func (d *Deployments) Scale(replicas int) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Deployment, d.address, d.namespace) + "/" + d.deployment + "/scale"
		header = map[string]string{
			"Content-Type":  "application/strategic-merge-patch+json",
			"Authorization": getToken(),
		}
		payload = fmt.Sprintf(`{"spec": {"replicas": %d}}`, replicas)
	)

	body, err := curl.Patch(url, header, []byte(payload), 5)
	if err != nil {
		log.Errorf("scale deployment: %s replicas: %d error: %s", d.deployment, replicas, err)
		return err
	}

	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("scale deployment: %s response json decode error: %s", d.deployment, err)
		return err
	}

	spec := resp["spec"].(map[string]interface{})
	if len(spec) != 0 && spec["replicas"].(float64) == float64(replicas) {
		log.Infof("scale deployment: %s replicas: %d success.", d.deployment, replicas)
	}
	return nil
}

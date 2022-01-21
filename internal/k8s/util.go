// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package k8s

import (
	"encoding/json"
	"fmt"

	"nautilus/pkg/g"
	"nautilus/pkg/log"
)

// NewDeployments deployment相关操作
func NewDeployments(namespace, deployment string) *Deployments {
	return &Deployments{
		namespace:  namespace,
		deployment: deployment,
	}
}

type Deployments struct {
	namespace  string
	deployment string
}

func (d *Deployments) Exist() bool {
	var (
		url    = fmt.Sprintf(g.Config().K8S.Deployment, d.namespace) + "/" + d.deployment
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := g.Get(url, header, 5)
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
		url    = fmt.Sprintf(g.Config().K8S.Deployment, d.namespace)
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := g.Post(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request create deployment api error: %s", err)
		return err
	}
	return response(body)
}

func (d *Deployments) Update(tpl string) error {
	var (
		url    = fmt.Sprintf(g.Config().K8S.Deployment, d.namespace) + "/" + d.deployment
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := g.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update deployment api error: %s", err)
		return err
	}
	return response(body)
}

func (d *Deployments) Scale(replicas int) error {
	var (
		url     = fmt.Sprintf(g.Config().K8S.Deployment, d.namespace) + "/" + d.deployment + "/scale"
		header  = map[string]string{"Content-Type": "application/strategic-merge-patch+json"}
		payload = fmt.Sprintf(`{"spec": {"replicas": %d}}`, replicas)
	)

	body, err := g.Patch(url, header, []byte(payload), 5)
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

// NewServices service相关操作
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
		url = fmt.Sprintf(g.Config().K8S.Service, s.namespace) + "/" + s.name
	)
	body, err := g.Get(url, nil, 5)
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
		url    = fmt.Sprintf(g.Config().K8S.Service, s.namespace)
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := g.Post(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request create service api error: %s", err)
		return err
	}
	return response(body)
}

func (s *Services) Update(tpl string) error {
	var (
		url    = fmt.Sprintf(g.Config().K8S.Service, s.namespace) + "/" + s.name
		header = map[string]string{"Content-Type": "application/json"}
	)
	body, err := g.Put(url, header, []byte(tpl), 5)
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

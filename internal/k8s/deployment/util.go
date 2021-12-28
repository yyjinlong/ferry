// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"encoding/json"
	"fmt"

	"ferry/pkg/g"
	"ferry/pkg/log"
)

func NewDeployments(namespace, deployment string) *Deployments {
	return &Deployments{}
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
	body, err := g.Get(url, header, nil, 5)
	if err != nil {
		log.Infof("check deployment: %s is not exist", deployment)
		return false
	}
	if err := d.result(body); err != nil {
		return false
	}
	log.Infof("check deplloyment: %s is exist", deployment)
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
	return d.result(body)
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
	return d.result(body)
}

func (d *Deployments) result(body string) error {
	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("response body json decode error: %s", err)
		return err
	}

	status, ok := resp["status"].(string)
	if ok && status == "Failure" {
		err := fmt.Errorf("%s", resp["message"].(string))
		log.Errorf("request deployment api failed: %s", err)
		return err
	}
	log.Info("deployment operate success")
	return nil
}

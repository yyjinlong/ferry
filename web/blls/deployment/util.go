// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"encoding/json"
	"fmt"

	"ferry/ops/g"
	"ferry/ops/log"
)

func newDeployments() *deployments {
	return &deployments{}
}

type deployments struct {
}

func (d *deployments) exist(namespace, deployment string) bool {
	url := fmt.Sprintf(g.Config().K8S.Deployment, namespace) + "/" + deployment
	header := map[string]string{"Content-Type": "application/json"}
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

func (d *deployments) create(namespace, tpl string) error {
	url := fmt.Sprintf(g.Config().K8S.Deployment, namespace)
	header := map[string]string{"Content-Type": "application/json"}
	body, err := g.Post(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request create deployment api error: %s", err)
		return err
	}
	return d.result(body)
}

func (d *deployments) update(namespace, deployment, tpl string) error {
	url := fmt.Sprintf(g.Config().K8S.Deployment, namespace) + "/" + deployment
	header := map[string]string{"Content-Type": "application/json"}
	body, err := g.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update deployment api error: %s", err)
		return err
	}
	return d.result(body)
}

func (d *deployments) result(body string) error {
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

func (d *deployments) scale(replicas int, namespace, deployment string) error {
	url := fmt.Sprintf(g.Config().K8S.Deployment, namespace) + "/" + deployment + "/scale"
	header := map[string]string{"Content-Type": "application/strategic-merge-patch+json"}
	payload := fmt.Sprintf(`{"spec": {"replicas": %d}}`, replicas)
	body, err := g.Patch(url, header, []byte(payload), 5)
	if err != nil {
		log.Errorf("scale deployment: %s replicas: %d error: %s", deployment, replicas, err)
		return err
	}

	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("scale response body json decode result error: %s", err)
		return err
	}

	spec := resp["spec"].(map[string]interface{})
	if len(spec) != 0 && spec["replicas"].(float64) == float64(replicas) {
		log.Infof("scale deployment: %s replicas: %d success.", deployment, replicas)
	}
	return nil
}

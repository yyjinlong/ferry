// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package exec

import (
	"encoding/json"
	"fmt"

	"nautilus/golib/log"
	"nautilus/pkg/config"
	"nautilus/pkg/model"
)

func getToken() string {
	return fmt.Sprintf("Bearer %s", config.Config().K8S.Token)
}

func getHeader() map[string]string {
	return map[string]string{
		"Authorization": getToken(),
		"Content-Type":  "application/json",
	}
}

func getAddress(namespace string) string {
	ns, err := model.GetNamespaceByName(namespace)
	if err != nil {
		log.Errorf("query namespace mapping address error: %s", err)
		return ""
	}
	return config.GetAddress(ns.Cluster)
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
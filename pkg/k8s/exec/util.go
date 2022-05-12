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

func getCluster(namespace string) string {
	ns, err := model.GetNamespaceByName(namespace)
	if err != nil {
		log.Errorf("query address error: %+v by namespace: %s", err, namespace)
		return ""
	}
	return ns.Cluster
}

func getAddress(cluster string) string {
	return config.GetAddress(cluster)
}

func getToken(cluster string) string {
	obj, err := model.GetCluster(cluster)
	if err != nil {
		log.Errorf("query token error: %+v by cluster: %s", err, cluster)
		return ""
	}
	return fmt.Sprintf("Bearer %s", obj.Token)
}

func getHeader(cluster string) map[string]string {
	return map[string]string{
		"Authorization": getToken(cluster),
		"Content-Type":  "application/json",
	}
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

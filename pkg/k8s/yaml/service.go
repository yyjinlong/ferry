// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package yaml

import (
	"encoding/json"
	"fmt"
)

type ServiceYaml struct {
	ServiceName   string
	ServiceID     int64
	Phase         string
	AppID         string
	ExposePort    int
	ContainerPort int
}

func (sy *ServiceYaml) Instance() (string, error) {
	/*
	  apiVersion:
	  kind:
	  metadata:
	  spec:
	*/
	controller := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata":   sy.metadata(),
		"spec":       sy.spec(),
	}
	config, err := json.Marshal(controller)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (sy *ServiceYaml) metadata() interface{} {
	return map[string]interface{}{
		"name": fmt.Sprintf("%s-%d-%s", sy.ServiceName, sy.ServiceID, sy.Phase),
	}
}

func (sy *ServiceYaml) spec() interface{} {
	/*
	  type:
	  selector:
	    ...
	  ports:
	    ...
	*/
	spec := make(map[string]interface{})
	spec["type"] = "ClusterIP"
	spec["selector"] = sy.selector()
	spec["ports"] = sy.ports()
	return spec
}

func (sy *ServiceYaml) selector() interface{} {
	return map[string]string{
		"appid": sy.AppID,
	}
}

func (sy *ServiceYaml) ports() interface{} {
	cluster := make(map[string]interface{})
	cluster["port"] = sy.ExposePort
	cluster["targetPort"] = sy.ContainerPort
	return []interface{}{cluster}
}

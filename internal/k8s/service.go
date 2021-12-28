// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package k8s

import (
	"encoding/json"
	"fmt"
)

type ServiceYaml struct {
	ServiceName   string
	ServiceID     int64
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
	type controller struct {
		apiVersion string
		kind       string
		metadata   interface{}
		spec       interface{}
	}
	ctl := controller{"v1", "Service", sy.metadata(), sy.spec()}

	config, err := json.Marshal(ctl)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (sy *ServiceYaml) metadata() interface{} {
	return map[string]interface{}{
		"name": fmt.Sprintf("%s-%d", sy.ServiceName, sy.ServiceID),
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

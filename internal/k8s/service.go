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
	serviceName   string
	serviceID     int64
	appid         string
	exposePort    int
	containerPort int
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

	config, err := json.Marshal(vip)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (sy *ServiceYaml) metadata() interface{} {
	return map[string]interface{}{
		"name": fmt.Sprintf("%s-%d", sy.serviceName, sy.serviceID),
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
		"appid": sy.appid,
	}
}

func (sy *ServiceYaml) ports() interface{} {
	cluster := make(map[string]interface{})
	cluster["port"] = sy.exposePort
	cluster["targetPort"] = sy.containerPort
	return []interface{}{cluster}
}

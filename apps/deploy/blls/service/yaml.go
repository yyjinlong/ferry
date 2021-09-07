// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package service

import (
	"encoding/json"
	"fmt"
)

type yaml struct {
	serviceName   string
	serviceID     int64
	appid         string
	exposePort    int
	containerPort int
}

func (y *yaml) instance() (string, error) {
	/*
	  apiVersion:
	  kind:
	  metadata:
	  spec:
	*/
	vip := make(map[string]interface{})
	vip["apiVersion"] = "v1"
	vip["kind"] = "Service"
	vip["metadata"] = y.metadata()
	vip["spec"] = y.spec()

	config, err := json.Marshal(vip)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (y *yaml) metadata() interface{} {
	return map[string]interface{}{
		"name": fmt.Sprintf("%s-%d", y.serviceName, y.serviceID),
	}
}

func (y *yaml) spec() interface{} {
	/*
	  type:
	  selector:
	    ...
	  ports:
	    ...
	*/
	spec := make(map[string]interface{})
	spec["type"] = "ClusterIP"
	spec["selector"] = y.selector()
	spec["ports"] = y.ports()
	return spec
}

func (y *yaml) selector() interface{} {
	return map[string]string{
		"appid": y.appid,
	}
}

func (y *yaml) ports() interface{} {
	cluster := make(map[string]interface{})
	cluster["port"] = y.exposePort
	cluster["targetPort"] = y.containerPort
	return []interface{}{cluster}
}

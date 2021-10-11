// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"ferry/ops/log"
	"ferry/tracer/handle"
)

func EndpointAdd(obj interface{}) {
	epEvent := handle.NewEndpointEvent(obj, "add")
	if !epEvent.IsValidService() {
		return
	}
	pipelineID, err := epEvent.Parse()
	if err != nil {
		log.Errorf("[add] parser pipeline id error: %s", err)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    epEvent.GetService(),
		"add":        epEvent.GetIPList(),
		"del":        make([]map[string]string, 0),
	}
	log.Infof("endpoint add result: %+v", result)
}

func EndpointUpdate(oldObj, newObj interface{}) {
	oldEvent := handle.NewEndpointEvent(oldObj, "update")
	if !oldEvent.IsValidService() {
		return
	}
	newEvent := handle.NewEndpointEvent(newObj, "update")
	if !newEvent.IsValidService() {
		return
	}
	pipelineID, err := newEvent.Parse()
	if err != nil {
		log.Errorf("[update] parser pipeline id error: %s", err)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    newEvent.GetService(),
		"add":        newEvent.GetIPList(),
		"del":        oldEvent.GetIPList(),
	}
	log.Infof("endpoint update result: %+v", result)
}

func EndpointDelete(obj interface{}) {
	epEvent := handle.NewEndpointEvent(obj, "delete")
	if !epEvent.IsValidService() {
		return
	}
	pipelineID, err := epEvent.Parse()
	if err != nil {
		log.Errorf("[delete] parser pipeline id error: %s", err)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    epEvent.GetService(),
		"add":        make([]map[string]string, 0),
		"del":        epEvent.GetIPList(),
	}
	log.Infof("endpoint delete result: %+v", result)
}

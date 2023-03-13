// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package yaml

import (
	"encoding/json"
)

type ConfigmapYaml struct {
	Namespace string
	Name      string
}

func (cy *ConfigmapYaml) Instance(data map[string]string) (string, error) {
	controller := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata":   cy.metadata(),
		"data":       data,
	}
	config, err := json.Marshal(controller)
	if err != nil {
		return "", nil
	}
	return string(config), nil
}

func (cy *ConfigmapYaml) metadata() map[string]string {
	return map[string]string{
		"name":      cy.Name,
		"namespace": cy.Namespace,
	}
}

// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"ferry/ops/base"
	"ferry/ops/db"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/ops/objects"
)

type Finish struct {
	pid int64
}

func (f *Finish) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := f.checkParam(c, r.RequestID); err != nil {
		return nil, err
	}
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": f.pid})

	pipelineObj, err := objects.GetPipelineInfo(f.pid)
	if err != nil {
		log.Errorf("get pipeline info error: %s", err)
		return nil, err
	}

	if !f.clearOld(pipelineObj) {
		return nil, fmt.Errorf("old group deployment scale to 0 failed")
	}

	if err := f.setOnline(pipelineObj); err != nil {
		return nil, err
	}
	return "", nil
}

func (f *Finish) checkParam(c *gin.Context, logid string) error {
	type params struct {
		ID int64 `form:"pipeline_id" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return err
	}
	f.pid = data.ID
	return nil
}

func (f *Finish) clearOld(pipelineObj *db.PipelineQuery) bool {
	namespace := pipelineObj.Namespace.Name
	service := pipelineObj.Service.Name

	// NOTE: 在确认时, 原有表记录的组则变为待下线组
	offlineGroup := pipelineObj.Service.OnlineGroup
	log.Infof("get current clear offline group: %s", offlineGroup)
	if offlineGroup == "none" {
		return true
	}

	for _, phase := range db.PHASE_NAME_LIST {
		deployment := objects.GetDeployment(f.pid, service, phase, offlineGroup)
		if f.isExist(namespace, deployment) {
			if err := f.scale(0, namespace, deployment); err != nil {
				log.Errorf("scale deployment: %s replicas: 0 error: %s", deployment, err)
				return false
			}
			log.Infof("scale deployment: %s replicas: 0 success", deployment)
		}
	}
	return true
}

func (f *Finish) isExist(namespace, deployment string) bool {
	url := fmt.Sprintf(g.Config().K8S.Deployment, namespace) + "/" + deployment
	header := map[string]string{"Content-Type": "application/json"}
	_, err := g.Get(url, header, nil, 5)
	if err != nil {
		log.Errorf("check deployment: %s is not exist. error: %s", deployment, err)
		return false
	}
	log.Infof("check deplloyment: %s is exist.", deployment)
	return true
}

func (f *Finish) scale(replicas int, namespace, deployment string) error {
	url := fmt.Sprintf(g.Config().K8S.Deployment, namespace) + "/" + deployment + "/scale"
	header := map[string]string{"Content-Type": "application/strategic-merge-patch+json"}
	payload := fmt.Sprintf(`{"spec": {"replicas": %d}}`, replicas)
	body, err := g.Patch(url, header, []byte(payload), 5)
	if err != nil {
		log.Errorf("scale deployment: %s replicas: %d error: %s", deployment, replicas, err)
		return err
	}

	log.Infof("scale deployment: %s response: %s", deployment, body)
	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("json decode http result error: %s", err)
		return err
	}

	spec := resp["spec"].(map[string]interface{})
	if len(spec) != 0 {
		log.Infof("scale deployment: %s spec result: %s", deployment, spec)
		if spec["replicas"].(float64) == float64(replicas) {
			log.Infof("scale deployment: %s replicas: %d success.", deployment, replicas)
		}
	}
	return nil
}

func (f *Finish) setOnline(pipelineObj *db.PipelineQuery) error {
	group := objects.GetDeployGroup(pipelineObj.Service.OnlineGroup)
	log.Infof("get current online group: %s", group)

	if err := objects.UpdateGroup(f.pid, pipelineObj.Service.Name, group); err != nil {
		log.Errorf("set current group: %s online error: %s", group, err)
	}
	log.Infof("set current group: %s online success.", group)
	return nil
}

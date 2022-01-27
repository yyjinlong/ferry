// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"nautilus/internal/model"
	"nautilus/internal/objects"
	"nautilus/pkg/base"
	"nautilus/pkg/g"
	"nautilus/pkg/log"
	"nautilus/pkg/mq"
)

type BuildImage struct{}

func (bi *BuildImage) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		ID      int64  `form:"pipeline_id" binding:"required"`
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return nil, err
	}

	var (
		pid      = data.ID
		service  = data.Service
		language string
	)
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": pid})

	pipeline, err := objects.GetPipeline(pid)
	if err != nil {
		return nil, fmt.Errorf(IMG_QUERY_PIPELINE_ERROR, err)
	}

	if g.Ini(pipeline.Status, []int{model.PLSuccess, model.PLFailed, model.PLRollbackSuccess, model.PLRollbackFailed, model.PLTerminate}) {
		return nil, fmt.Errorf(IMG_BUILD_FINISHED)
	}

	updateList, err := objects.FindUpdateInfo(pid)
	if err != nil {
		log.Errorf("find pipeline update info error: %s", err)
		return nil, fmt.Errorf(IMG_QUERY_UPDATE_ERROR, err)
	}

	builds := make([]map[string]string, 0)
	for _, item := range updateList {
		language = item.CodeModule.Language

		tagInfo := map[string]string{
			"module": item.CodeModule.Name,
			"repo":   item.CodeModule.ReposAddr,
			"tag":    item.PipelineUpdate.CodeTag,
		}
		builds = append(builds, tagInfo)
	}

	image := map[string]interface{}{
		"pid":     pid,
		"type":    language,
		"service": service,
		"build":   builds,
	}
	body, err := json.Marshal(image)
	if err != nil {
		return nil, fmt.Errorf(IMG_BUILD_PARAM_ENCODE_ERROR, err)
	}
	log.Infof("publish build image body: %s", string(body))

	mqConf := g.Config().RabbitMQ
	rmq := mq.NewRabbitMQ(mqConf.Address, mqConf.Exchange, mqConf.Queue, mqConf.RoutingKey)
	rmq.Publish(string(body))
	log.Infof("publish build image info to rmq success.")
	return "", nil
}

// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"ferry/pkg/g"
	"ferry/pkg/log"
	"ferry/pkg/mq"
	"ferry/server/base"
	"ferry/server/objects"
)

type BuildImage struct{}

func (bi *BuildImage) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		ID int64 `form:"pipeline_id" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return nil, err
	}

	var (
		pid      = data.ID
		service  string
		language string
	)
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": pid})

	updateList, err := objects.FindUpdateInfo(pid)
	if errors.Is(err, objects.NotFound) {
		return nil, fmt.Errorf("pipeline_id: %d 不存在!", pid)
	} else if err != nil {
		log.Errorf("find pipeline update info error: %s", err)
		return nil, err
	}

	builds := make([]map[string]string, 0)
	for _, item := range updateList {
		service = item.Service.Name
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
		return nil, err
	}
	log.Infof("publish build image body: %s", string(body))

	mqConf := g.Config().RabbitMQ
	rmq := mq.NewRabbitMQ(mqConf.Address, mqConf.Exchange, mqConf.Queue, mqConf.RoutingKey)
	rmq.Publish(string(body))
	log.Infof("publish build image info to rmq success.")
	return "", nil
}

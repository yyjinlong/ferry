// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package pipeline

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"ferry/ops/base"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/ops/mq"
	"ferry/ops/objects"
)

type BuildImage struct {
	pid int64
}

func (bi *BuildImage) validate(c *gin.Context) error {
	type params struct {
		ID int64 `form:"pipeline_id" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return err
	}
	bi.pid = data.ID
	return nil
}

func (bi *BuildImage) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := bi.validate(c); err != nil {
		return nil, err
	}
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": bi.pid})

	updateList, err := objects.FindUpdateInfo(bi.pid)
	if err != nil {
		log.Errorf("find pipeline update info error: %s", err)
		return nil, err
	}

	var (
		service  string
		language string
	)

	builds := make([]map[string]string, 0)
	for _, item := range updateList {
		service = item.Service.Name
		language = item.CodeModule.Language

		tagInfo := make(map[string]string)
		tagInfo["module"] = item.CodeModule.Name
		tagInfo["repo"] = item.CodeModule.ReposAddr
		tagInfo["tag"] = item.PipelineUpdate.CodeTag
		builds = append(builds, tagInfo)
	}

	image := make(map[string]interface{})
	image["pid"] = bi.pid
	image["type"] = language
	image["service"] = service
	image["build"] = builds
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

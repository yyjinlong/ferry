// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
	"nautilus/pkg/util/rmq"
)

var (
	msgChan = make(chan Image)
)

type receiver struct{}

func (r *receiver) Consumer(body []byte) error {
	var data Image
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("consume rabbitmq json decode failed: %s", err)
		return err
	}
	msgChan <- data
	return nil
}

func ListenMQ() {
	mq, err := rmq.NewRabbitMQ(
		config.Config().RabbitMQ.Address,
		config.Config().RabbitMQ.Exchange,
		config.Config().RabbitMQ.Queue,
		config.Config().RabbitMQ.RoutingKey)
	if err != nil {
		log.Panicf("boot connect amqp failed: %s", err)
	}
	mq.Consume(&receiver{})
}

func HandleMsg() {
	for {
		select {
		case data := <-msgChan:
			go worker(data)
		}
	}
}

func getCurPath() string {
	_, curPath, _, _ := runtime.Caller(1)
	return curPath
}

func worker(data Image) {
	var (
		pid       = data.PID
		service   = data.Service
		buildPath = filepath.Join(config.Config().Image.Dir, service, strconv.FormatInt(pid, 10))
		appPath   = filepath.Dir(filepath.Dir(getCurPath()))
		codePath  = filepath.Join(buildPath, "code")
		imageURL  = fmt.Sprintf("%s/%s", config.Config().Image.Registry, service)
		imageTag  = fmt.Sprintf("v-%s", time.Now().Format("20060102_150405"))
		targetURL = fmt.Sprintf("%s:%s", imageURL, imageTag)
	)

	util.Mkdir(buildPath) // 构建路径: 主路径/服务/上线单ID
	util.Mkdir(codePath)  // 代码路径: 主路径/服务/上线单ID/code

	for _, item := range data.Build {
		if err := downloadCode(item.Module, item.Repo, item.Tag, codePath); err != nil {
			log.Errorf("download code failed: %+v", err)
			return
		}
	}

	if err := compile(data.Type); err != nil {
		log.Errorf("compile code failed: %+v", err)
		return
	}

	p := &pipeline{
		pid:       data.PID,
		service:   data.Service,
		appPath:   appPath,
		buildPath: buildPath,
		codePath:  codePath,
		imageURL:  imageURL,
		imageTag:  imageTag,
		targetURL: targetURL,
	}

	if !p.copyDockerfile() {
		return
	}
	if !p.dockerBuild() {
		return
	}
	if !p.dockerPush() {
		return
	}

	if err := model.UpdateImage(p.pid, p.imageURL, p.imageTag); err != nil {
		log.Errorf("write image info to db error: %s", err)
		return
	}
	log.Infof("write relase image: %s to database success", p.targetURL)
}

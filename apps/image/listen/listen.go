// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"encoding/json"

	"ferry/apps/image/build"
	"ferry/ops/base"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/ops/mq"
)

var (
	pyChan = make(chan base.Image)
	goChan = make(chan base.Image)
)

func BuildImage() {
	done := make(chan int)

	go listenMQ()
	go handlePy()
	go handleGo()

	<-done
}

type mirror struct {
}

func (m *mirror) Consumer(body []byte) error {
	var data base.Image
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("consume mq json decode failed: %s", err)
		return err
	}
	log.InitFields(log.Fields{
		"logid": g.UniqueID(), "pid": data.PID, "service": data.Service, "type": data.Type})

	switch data.Type {
	case base.PYTHON:
		pyChan <- data
	case base.GOLANG:
		goChan <- data
	}
	return nil
}

func listenMQ() {
	mqConf := g.Config().RabbitMQ
	rmq := mq.NewRabbitMQ(mqConf.Address, mqConf.Exchange, mqConf.Queue, mqConf.RoutingKey)
	rmq.Consume(&mirror{})
}

func handlePy() {
	for {
		select {
		case data := <-pyChan:
			go build.Python(data)
		}
	}
}

func handleGo() {
	for {
		select {
		case data := <-goChan:
			go build.Golang(data)
		}
	}
}

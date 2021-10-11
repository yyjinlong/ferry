// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package build

import (
	"encoding/json"

	"ferry/imager/model"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/ops/mq"
)

var (
	pyChan = make(chan model.Image)
	goChan = make(chan model.Image)
)

func ListenImage() {
	go listenMQ()
	go handlePy()
	go handleGo()

	done := make(chan int)
	<-done
}

type mirror struct {
}

func (m *mirror) Consumer(body []byte) error {
	var data model.Image
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("consume mq json decode failed: %s", err)
		return err
	}
	log.InitFields(log.Fields{
		"logid": g.UniqueID(), "pid": data.PID, "service": data.Service, "type": data.Type})

	switch data.Type {
	case model.PYTHON:
		pyChan <- data
	case model.GOLANG:
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
			go BuildPython(data)
		}
	}
}

func handleGo() {
	for {
		select {
		case data := <-goChan:
			go BuildGolang(data)
		}
	}
}

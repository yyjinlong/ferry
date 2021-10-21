// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package mirror

import (
	"encoding/json"

	"ferry/pkg/g"
	"ferry/pkg/log"
	"ferry/pkg/mq"
)

var (
	pyChan = make(chan Image)
	goChan = make(chan Image)
)

type mirror struct {
}

func (m *mirror) Consumer(body []byte) error {
	var data Image
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("consume mq json decode failed: %s", err)
		return err
	}
	log.InitFields(log.Fields{
		"logid": g.UniqueID(), "pid": data.PID, "service": data.Service, "type": data.Type})

	switch data.Type {
	case PYTHON:
		pyChan <- data
	case GOLANG:
		goChan <- data
	}
	return nil
}

func ListenMQ() {
	mqConf := g.Config().RabbitMQ
	rmq := mq.NewRabbitMQ(mqConf.Address, mqConf.Exchange, mqConf.Queue, mqConf.RoutingKey)
	rmq.Consume(&mirror{})
}

func HandlePy() {
	for {
		select {
		case data := <-pyChan:
			go Execute(data)
		}
	}
}

func HandleGo() {
	for {
		select {
		case data := <-goChan:
			go Execute(data)
		}
	}
}

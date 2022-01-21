// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package image

import (
	"encoding/json"

	"nautilus/pkg/g"
	"nautilus/pkg/log"
	"nautilus/pkg/mq"
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
	log.InitFields(log.Fields{
		"logid": g.UniqueID(), "pid": data.PID, "service": data.Service, "type": "image"})
	msgChan <- data
	return nil
}

func ListenMQ() {
	mqConf := g.Config().RabbitMQ
	rmq := mq.NewRabbitMQ(mqConf.Address, mqConf.Exchange, mqConf.Queue, mqConf.RoutingKey)
	rmq.Consume(&receiver{})
}

func HandleMsg() {
	for {
		select {
		case data := <-msgChan:
			go worker(data)
		}
	}
}

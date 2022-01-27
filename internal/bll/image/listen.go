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
	rmq, err := mq.NewRabbitMQ(
		g.Config().RabbitMQ.Address,
		g.Config().RabbitMQ.Exchange,
		g.Config().RabbitMQ.Queue,
		g.Config().RabbitMQ.RoutingKey)
	if err != nil {
		log.Panicf("boot connect amqp failed: %s", err)
	}
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

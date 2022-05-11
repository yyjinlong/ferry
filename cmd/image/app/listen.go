// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"encoding/json"

	"nautilus/golib/log"
	"nautilus/golib/rmq"
	"nautilus/pkg/config"
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
	log.InitFields(log.Fields{"pid": data.PID, "service": data.Service})
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

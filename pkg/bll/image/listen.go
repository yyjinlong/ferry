// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package image

import (
	"encoding/json"

	"github.com/yyjinlong/golib/log"
	"github.com/yyjinlong/golib/rmq"

	"nautilus/pkg/cfg"
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
		cfg.Config().RabbitMQ.Address,
		cfg.Config().RabbitMQ.Exchange,
		cfg.Config().RabbitMQ.Queue,
		cfg.Config().RabbitMQ.RoutingKey)
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

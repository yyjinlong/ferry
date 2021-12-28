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
	msgChan = make(chan Image)
)

type mirror struct{}

func (m *mirror) Consumer(body []byte) error {
	var data Image
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("consume mq json decode failed: %s", err)
		return err
	}
	log.InitFields(log.Fields{
		"logid": g.UniqueID(), "pid": data.PID, "service": data.Service, "type": "mirror"})
	msgChan <- data
	return nil
}

func ListenMQ() {
	mqConf := g.Config().RabbitMQ
	rmq := mq.NewRabbitMQ(mqConf.Address, mqConf.Exchange, mqConf.Queue, mqConf.RoutingKey)
	rmq.Consume(&mirror{})
}

func HandleMsg() {
	for {
		select {
		case data := <-msgChan:
			go worker(data)
		}
	}
}

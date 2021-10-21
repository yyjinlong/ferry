// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package mq

import (
	"sync"

	"github.com/streadway/amqp"

	"ferry/pkg/log"
)

var (
	once sync.Once
)

type Receiver interface {
	Consumer([]byte) error
}

func NewRabbitMQ(addr, exchange, queue, routingKey string) *RabbitMQ {
	rmq := &RabbitMQ{
		addr:       addr,
		exchange:   exchange,
		queue:      queue,
		routingKey: routingKey,
	}
	rmq.Connect()
	once.Do(func() {
		rmq.ExchangeDeclare()
		rmq.QueueDeclare()
		rmq.QueueBind()
	})
	return rmq
}

type RabbitMQ struct {
	addr       string
	exchange   string
	queue      string
	routingKey string
	conn       *amqp.Connection
	channel    *amqp.Channel
}

func (r *RabbitMQ) Connect() {
	var err error
	if r.conn, err = amqp.Dial(r.addr); err != nil {
		log.Panicf("connect mq: %s failed: %s", r.addr, err)
	}

	if r.channel, err = r.conn.Channel(); err != nil {
		log.Panicf("open a channel failed: %s", err)
	}
}

func (r *RabbitMQ) Close() {
	if err := r.channel.Close(); err != nil {
		log.Errorf("close channel failed: %s", err)
	}

	if err := r.conn.Close(); err != nil {
		log.Errorf("colse connect failed: %s", err)
	}
}

func (r *RabbitMQ) ExchangeDeclare() {
	if err := r.channel.ExchangeDeclare(
		r.exchange,
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	); err != nil {
		log.Errorf("declare an exchange failed: %s", err)
	}
}

func (r *RabbitMQ) QueueDeclare() {
	if _, err := r.channel.QueueDeclare(
		r.queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		true,  // no-wait
		nil,   // arguments
	); err != nil {
		log.Errorf("declare a queue failed: %s", err)
	}
}

func (r *RabbitMQ) QueueBind() {
	if err := r.channel.QueueBind(
		r.queue,
		r.routingKey,
		r.exchange,
		true, // no-wait
		nil,  // arguments
	); err != nil {
		log.Errorf("bind queue to exchange failed: %s", err)
	}
}

func (r *RabbitMQ) Consume(receiver Receiver) {
	defer r.Close()
	msgs, err := r.channel.Consume(
		r.queue,
		"",    // consumer
		false, // auto ack
		false, // exclusive
		false, // no local
		false, // no wait
		nil,   // args
	)
	if err != nil {
		log.Errorf("register a consumer failed: %s", err)
		return
	}

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			if err := receiver.Consumer(msg.Body); err != nil {
				msg.Ack(true)
				return
			}
			msg.Ack(false)
		}
	}()

	select {
	case <-forever:
		log.Infof("stop receive message.")
	}
}

func (r *RabbitMQ) Publish(body string) {
	err := r.channel.Publish(
		r.exchange,
		r.routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		log.Errorf("publish message to rabbitmq failed: %s", err)
		return
	}
}

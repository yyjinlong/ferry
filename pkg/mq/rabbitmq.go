// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package mq

import (
	"sync"
	"time"

	"github.com/streadway/amqp"

	"nautilus/pkg/log"
)

var (
	once sync.Once
)

type Receiver interface {
	Consumer([]byte) error
}

func NewRabbitMQ(addr, exchange, queue, routingKey string) (*RabbitMQ, error) {
	rmq := &RabbitMQ{
		addr:         addr,
		exchange:     exchange,
		queue:        queue,
		routingKey:   routingKey,
		connChangeCh: make(chan struct{}, 1),
	}
	if err := rmq.Connect(); err != nil {
		return nil, err
	}
	once.Do(func() {
		rmq.ExchangeDeclare()
		rmq.QueueDeclare()
		rmq.QueueBind()
	})
	return rmq, nil
}

type RabbitMQ struct {
	addr          string
	exchange      string
	queue         string
	routingKey    string
	conn          *amqp.Connection
	channel       *amqp.Channel
	connNotify    chan *amqp.Error
	channelNotify chan *amqp.Error
	isConnected   bool
	connChangeCh  chan struct{} // 连接改变
}

func (r *RabbitMQ) Connect() error {
	var err error
	if r.conn, err = amqp.Dial(r.addr); err != nil {
		log.Errorf("connect mq: %s failed: %s", r.addr, err)
		return err
	}

	if r.channel, err = r.conn.Channel(); err != nil {
		log.Errorf("open a channel failed: %s", err)
		return err
	}
	r.isConnected = true

	connErrCh := make(chan *amqp.Error, 1)
	r.connNotify = r.conn.NotifyClose(connErrCh)

	chanErrCh := make(chan *amqp.Error, 1)
	r.channelNotify = r.channel.NotifyClose(chanErrCh)

	r.connChangeCh <- struct{}{}
	return nil
}

func (r *RabbitMQ) Close() {
	if err := r.channel.Close(); err != nil {
		log.Errorf("close channel failed: %s", err)
	}

	if err := r.conn.Close(); err != nil {
		log.Errorf("colse connect failed: %s", err)
	}
}

func (r *RabbitMQ) Reconnect() {
	for {
		if !r.isConnected {
			connected := false
			for i := 0; !connected; i++ {
				if err := r.Connect(); err != nil {
					log.Errorf("retry connect failed! count: %d", i)
					time.Sleep(2 * time.Second)
					continue
				}
				connected = true
				log.Infof("reconnect rabbitmq success! count: %d", i)
			}
		}

		select {
		case err := <-r.channelNotify:
			log.Errorf("channel close notify: %+v", err)
			r.isConnected = false
		case err := <-r.connNotify:
			log.Errorf("connect close notify: %+v", err)
			r.isConnected = false
		}
		log.Infof("reconnect fetch isConnected: %t", r.isConnected)
		time.Sleep(2 * time.Second)
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
	var (
		msgs <-chan amqp.Delivery
		err  error
	)

	go r.Reconnect()

	for {
		select {
		case <-r.connChangeCh:
			log.Infof("event 'Connect Change' triggered.")
			msgs, err = r.channel.Consume(
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
				break
			}

		default:
			if !r.isConnected || msgs == nil {
				// 没有连接, 必须等待连接事件, 以便其连接上后才进行消息消费
				time.Sleep(1 * time.Second)
				break
			}
			for msg := range msgs {
				if err := receiver.Consumer(msg.Body); err != nil {
					msg.Ack(true)
					continue
				}
				msg.Ack(false)
			}
		}
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

package openmock

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	amqpOnce    sync.Once
	amqpChannel *amqp.Channel
)

const rabbitContentType = "application/octet-stream"

func publishToAMQP(amqpURL string, exchange string, routingKey string, payload string) {
	var err error

	defer func() {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"exchange":   exchange,
			"routingKey": routingKey,
			"payload":    payload,
		}).Info("try to publish to amqp")
	}()

	ch := getAMQPChannel(amqpURL)
	if err := prepareChannel(ch, exchange, routingKey, ""); err != nil {
		logrus.Errorf("%s: %s", "failed to prepare the channel for amqp", err)
		return
	}

	err = ch.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: rabbitContentType,
			Body:        []byte(payload),
		})
	if err != nil {
		logrus.Errorf("%s: %s", "failed to publish message", err)
		return
	}
}

func getAMQPChannel(amqpURL string) *amqp.Channel {
	amqpOnce.Do(func() {
		conn, err := amqp.Dial(amqpURL)
		if err != nil {
			logrus.Fatalf("%s: %s", "failed to connect ot AMQP", err)
		}

		ch, err := conn.Channel()
		if err != nil {
			logrus.Fatalf("%s: %s", "failed to open a channel", err)
		}
		amqpChannel = ch
	})
	return amqpChannel
}

func (om *OpenMock) startAMQP() {
	mocks := om.repo.AMQPMocks
	for amqp, ms := range mocks {
		func(amqp ExpectAMQP, ms MocksArray) {
			ch := getAMQPChannel(om.AMQPURL)
			if err := prepareChannel(ch, amqp.Exchange, amqp.RoutingKey, amqp.Queue); err != nil {
				logrus.Errorf("%s: %s", "failed to prepare the channel for amqp", err)
				return
			}

			msgs, err := ch.Consume(
				amqp.Queue, // queue
				"",         // consumer unique name, if "", it will be randomly generated
				true,       // auto-ack
				false,      // exclusive
				false,      // no-local
				false,      // no-wait
				nil,        // args
			)
			if err != nil {
				logrus.Errorf("%s: %s", "failed to consume from a queue", err)
				return
			}
			go func() {
				for msg := range msgs {
					logrus.WithFields(logrus.Fields{
						"amqp_msg":    string(msg.Body),
						"exchange":    msg.Exchange,
						"routing_key": msg.RoutingKey,
					}).Info()
					ms.DoActions(&Context{
						AMQPPayload: string(msg.Body),
						om:          om,
					})
				}
			}()
		}(amqp, ms)
	}
}

func prepareChannel(ch *amqp.Channel, exchange string, routingKey string, queue string) error {
	if queue == "" {
		queue = routingKey
	}

	err := ch.ExchangeDeclare(
		exchange, // exchange name
		"topic",  // exchange type, by default set as "topic"
		true,     // durable
		false, false, false, nil)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		queue, // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	return nil
}

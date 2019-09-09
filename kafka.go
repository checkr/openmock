package openmock

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/sirupsen/logrus"
)

// KafkaPipelineFunc defines pipeline functions
// For example, decode/encode messages
type KafkaPipelineFunc func(c Context, in []byte) (out []byte, err error)

// DefaultPipelineFunc directly outputs the in bytes
var DefaultPipelineFunc = func(c Context, in []byte) (out []byte, err error) {
	return in, nil
}

type kafkaClient struct {
	clientID  string
	brokers   []string
	producer  sarama.SyncProducer
	consumers []*cluster.Consumer

	cFunc KafkaPipelineFunc
	pFunc KafkaPipelineFunc
}

func (kc *kafkaClient) sendMessage(topic string, bytes []byte) (err error) {
	var out []byte
	defer func() {
		logrus.WithFields(logrus.Fields{
			"err":         err,
			"topic":       topic,
			"payload":     string(bytes),
			"out_message": string(out),
		}).Info("try to publish to kafka")
	}()

	c := Context{
		KafkaTopic:   topic,
		KafkaPayload: string(bytes),
	}
	out, err = kc.pFunc(c, bytes)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(out),
	}
	_, _, err = kc.producer.SendMessage(msg)
	return err
}

func (kc *kafkaClient) close() error {
	if kc == nil {
		return nil
	}

	for _, consumer := range kc.consumers {
		if err := consumer.Close(); err != nil {
			return err
		}
	}
	if err := kc.producer.Close(); err != nil {
		return err
	}
	return nil
}

func (om *OpenMock) configKafka() error {
	producer, err := sarama.NewSyncProducer(om.KafkaSeedBrokers, nil)
	if err != nil {
		return err
	}
	kc := &kafkaClient{
		clientID: om.KafkaClientID,
		brokers:  om.KafkaSeedBrokers,
		producer: producer,
		cFunc:    DefaultPipelineFunc,
		pFunc:    DefaultPipelineFunc,
	}
	if om.KafkaConsumePipelineFunc != nil {
		kc.cFunc = om.KafkaConsumePipelineFunc
	}
	if om.KafkaPublishPipelineFunc != nil {
		kc.pFunc = om.KafkaPublishPipelineFunc
	}
	om.kafkaClient = kc
	return nil
}

func (om *OpenMock) startKafka() {
	if err := om.configKafka(); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Errorf("failed to config kafka")
		return
	}
	for kafka, ms := range om.repo.KafkaMocks {
		go func(kafka ExpectKafka, ms MocksArray) {
			consumer, err := cluster.NewConsumer(
				om.kafkaClient.brokers,
				om.kafkaClient.clientID,
				[]string{kafka.Topic},
				nil,
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":   err,
					"topic": kafka.Topic,
				}).Errorf("failed to create a consumer")
				return
			}
			logrus.Infof("consumer started for topic:%s", kafka.Topic)
			om.kafkaClient.consumers = append(om.kafkaClient.consumers, consumer)

			//nolint:gosimple
			for {
				select {
				case msg, ok := <-consumer.Messages():
					if ok {
						c := Context{
							KafkaTopic:   msg.Topic,
							KafkaPayload: string(msg.Value),
							om:           om,
						}
						payload, err := om.kafkaClient.cFunc(c, msg.Value)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"err":   err,
								"topic": kafka.Topic,
							}).Errorf("failed to decode msg when consume the message")
							return
						}
						c.KafkaPayload = string(payload)
						if err := ms.DoActions(c); err != nil {
							logrus.WithFields(logrus.Fields{
								"err":   err,
								"topic": kafka.Topic,
							}).Errorf("failed to do actions inside kafka consumer")
							return
						}
						consumer.MarkOffset(msg, "")
					}
				}
			}
		}(kafka, ms)
	}
}

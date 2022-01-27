package openmock

import (
	"crypto/tls"

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

func (om *OpenMock) saramaConsumerConfig() (config *cluster.Config, seedBrokers []string) {
	config = cluster.NewConfig()

	shouldEnableTLS := om.KafkaTLSConsumerEnabled || om.KafkaTLSEnabled
	if shouldEnableTLS {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{}
	}

	saslUsername := om.KafkaSaslUsername
	if om.KafkaSaslConsumerUsername != "" {
		saslUsername = om.KafkaSaslConsumerUsername
	}

	saslPassword := om.KafkaSaslPassword
	if om.KafkaSaslConsumerPassword != "" {
		saslPassword = om.KafkaSaslConsumerPassword
	}

	if saslUsername != "" && saslPassword != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = saslUsername
		config.Net.SASL.Password = saslPassword
	}

	seedBrokers = om.KafkaSeedBrokers
	if len(om.KafkaConsumerSeedBrokers) != 0 {
		seedBrokers = om.KafkaConsumerSeedBrokers
	}

	return config, seedBrokers
}

func (om *OpenMock) saramaProducerConfig() (config *sarama.Config, seedBrokers []string) {
	config = sarama.NewConfig()

	shouldEnableTLS := om.KafkaTLSProducerEnabled || om.KafkaTLSEnabled
	if shouldEnableTLS {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{}
	}

	saslUsername := om.KafkaSaslUsername
	if om.KafkaSaslProducerUsername != "" {
		saslUsername = om.KafkaSaslProducerUsername
	}

	saslPassword := om.KafkaSaslPassword
	if om.KafkaSaslProducerPassword != "" {
		saslPassword = om.KafkaSaslProducerPassword
	}

	if saslUsername != "" && saslPassword != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = saslUsername
		config.Net.SASL.Password = saslPassword
	}

	seedBrokers = om.KafkaSeedBrokers
	if len(om.KafkaProducerSeedBrokers) != 0 {
		seedBrokers = om.KafkaProducerSeedBrokers
	}

	// Required for sync producer
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	return config, seedBrokers
}

func (om *OpenMock) configKafka() error {
	config, seedBrokers := om.saramaProducerConfig()
	producer, err := sarama.NewSyncProducer(seedBrokers, config)
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
		}).Fatal("failed to config kafka")
		return
	}
	for kafka, ms := range om.repo.KafkaMocks {
		go func(kafka ExpectKafka, ms MocksArray) {
			config, brokers := om.saramaConsumerConfig()
			consumer, err := cluster.NewConsumer(
				brokers,
				om.kafkaClient.clientID,
				[]string{kafka.Topic},
				config,
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":   err,
					"topic": kafka.Topic,
				}).Fatal("failed to create a consumer")
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

						newOmLogger(c).WithFields(logrus.Fields{
							"topic":   msg.Topic,
							"payload": c.KafkaPayload,
						}).Info("start_consuming_message")

						ms.DoActions(c)
						consumer.MarkOffset(msg, "")
					}
				}
			}
		}(kafka, ms)
	}
}

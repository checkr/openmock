package openmock

import (
	"github.com/sirupsen/logrus"
)

// OpenMock holds all the configuration of running openmock
type OpenMock struct {
	// HTTP host and port to serve http mocks
	HTTPEnabled bool
	HTTPPort    int
	HTTPHost    string

	// Kafka Settings
	KafkaEnabled             bool
	KafkaClientID            string
	KafkaSeedBrokers         []string
	KafkaConsumePipelineFunc KafkaPipelineFunc
	KafkaPublishPipelineFunc KafkaPipelineFunc

	// AMQP URL
	AMQPEnabled bool
	AMQPURL     string

	// The templates directory to load the YAML files
	// The dir is relative to the runtime binary
	TemplatesDir string

	// Redis Key Value Store
	RedisType string
	RedisURL  string

	// Prviates
	repo        *MockRepo
	kafkaClient *kafkaClient
	redis       RedisDoer
}

// Start starts the openmock
func (om *OpenMock) Start() {
	err := om.Load()
	if err != nil {
		logrus.Fatalf("%s: %s", "failed to load yaml templates for mocks", err)
	}
	om.SetRedis()

	if om.HTTPEnabled {
		go om.startHTTP()
	}
	if om.KafkaEnabled {
		go om.startKafka()
	}
	if om.AMQPEnabled {
		go om.startAMQP()
	}

	select {}
}

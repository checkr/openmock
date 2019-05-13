package openmock

import (
	"os"

	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
	"github.com/teamwork/reload"
)

// OpenMock holds all the configuration of running openmock
type OpenMock struct {
	// Env configuration
	LogLevel         string   `env:"OPENMOCK_LOG_LEVEL" envDefault:"info"`
	TemplatesDir     string   `env:"OPENMOCK_TEMPLATES_DIR" envDefault:"./templates"`
	HTTPEnabled      bool     `env:"OPENMOCK_HTTP_ENABLED" envDefault:"true"`
	HTTPPort         int      `env:"OPENMOCK_HTTP_PORT" envDefault:"9999"`
	HTTPHost         string   `env:"OPENMOCK_HTTP_HOST" envDefault:"0.0.0.0"`
	AdminHTTPEnabled bool     `env:"OPENMOCK_ADMIN_HTTP_ENABLED" envDefault:"true"`
	AdminHTTPPort    int      `env:"OPENMOCK_ADMIN_HTTP_PORT" envDefault:"9998"`
	AdminHTTPHost    string   `env:"OPENMOCK_ADMIN_HTTP_HOST" envDefault:"0.0.0.0"`
	KafkaEnabled     bool     `env:"OPENMOCK_KAFKA_ENABLED" envDefault:"false"`
	KafkaClientID    string   `env:"OPENMOCK_KAFKA_CLIENT_ID" envDefault:"openmock"`
	KafkaSeedBrokers []string `env:"OPENMOCK_KAFKA_SEED_BROKERS" envDefault:"kafka:9092,localhost:9092" envSeparator:","`
	AMQPEnabled      bool     `env:"OPENMOCK_AMQP_ENABLED" envDefault:"false"`
	AMQPURL          string   `env:"OPENMOCK_AMQP_URL" envDefault:"amqp://guest:guest@rabbitmq:5672"`
	RedisType        string   `env:"OPENMOCK_REDIS_TYPE" envDefault:"memory"`
	RedisURL         string   `env:"OPENMOCK_REDIS_URL" envDefault:"redis://redis:6379"`

	// Customized pipeline functions
	KafkaConsumePipelineFunc KafkaPipelineFunc
	KafkaPublishPipelineFunc KafkaPipelineFunc

	// Prviates
	repo        *MockRepo
	kafkaClient *kafkaClient
	redis       RedisDoer
}

// ParseEnv loads env vars into the openmock struct
func (om *OpenMock) ParseEnv() {
	env.Parse(om)
}

func (om *OpenMock) setupLogrus() {
	l, err := logrus.ParseLevel(om.LogLevel)
	if err != nil {
		logrus.WithField("err", err).Fatalf("failed to set logrus level:%s", om.LogLevel)
	}
	logrus.SetLevel(l)
	logrus.SetOutput(os.Stdout)
}

// Start starts the openmock
func (om *OpenMock) Start() {
	om.setupLogrus()
	om.SetRedis()
	om.StartAdmin()

	err := om.Load()
	if err != nil {
		logrus.Fatalf("%s: %s", "failed to load yaml templates for mocks", err)
	}

	if om.HTTPEnabled {
		go om.startHTTP()
	}
	if om.KafkaEnabled {
		go om.startKafka()
	}
	if om.AMQPEnabled {
		go om.startAMQP()
	}

	go func() {
		err := reload.Do(logrus.Infof, reload.Dir(om.TemplatesDir, reload.Exec))
		if err != nil {
			logrus.Fatal(err)
		}
	}()
	select {}
}

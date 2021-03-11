package openmock

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/goombaio/orderedmap"
	"github.com/sirupsen/logrus"
	"github.com/teamwork/reload"
)

// OpenMock holds all the configuration of running openmock
type OpenMock struct {
	LogLevel              string `env:"OPENMOCK_LOG_LEVEL" envDefault:"info"`
	TemplatesDir          string `env:"OPENMOCK_TEMPLATES_DIR" envDefault:"./templates"`
	TemplatesDirHotReload bool   `env:"OPENMOCK_TEMPLATES_DIR_HOT_RELOAD" envDefault:"true"`

	// HTTP channel
	HTTPEnabled bool   `env:"OPENMOCK_HTTP_ENABLED" envDefault:"true"`
	HTTPPort    int    `env:"OPENMOCK_HTTP_PORT" envDefault:"9999"`
	HTTPHost    string `env:"OPENMOCK_HTTP_HOST" envDefault:"0.0.0.0"`
	CorsEnabled bool   `env:"OPENMOCK_CORS_ENABLED" envDefault:"false"`

	// Admin channel
	AdminHTTPEnabled bool   `env:"OPENMOCK_ADMIN_HTTP_ENABLED" envDefault:"true"`
	AdminHTTPPort    int    `env:"OPENMOCK_ADMIN_HTTP_PORT" envDefault:"9998"`
	AdminHTTPHost    string `env:"OPENMOCK_ADMIN_HTTP_HOST" envDefault:"0.0.0.0"`

	// Kafka channel
	KafkaEnabled     bool     `env:"OPENMOCK_KAFKA_ENABLED" envDefault:"false"`
	KafkaClientID    string   `env:"OPENMOCK_KAFKA_CLIENT_ID" envDefault:"openmock"`
	KafkaSeedBrokers []string `env:"OPENMOCK_KAFKA_SEED_BROKERS" envDefault:"kafka:9092" envSeparator:","`

	// AMQP channel
	AMQPEnabled bool   `env:"OPENMOCK_AMQP_ENABLED" envDefault:"false"`
	AMQPURL     string `env:"OPENMOCK_AMQP_URL" envDefault:"amqp://guest:guest@rabbitmq:5672"`

	// Redis configuration for admin's ephemeral storage or redis commands
	RedisType string `env:"OPENMOCK_REDIS_TYPE" envDefault:"memory"`
	RedisURL  string `env:"OPENMOCK_REDIS_URL" envDefault:"redis://redis:6379"`

	// GRPC channel
	GRPCEnabled bool   `env:"OPENMOCK_GRPC_ENABLED" envDefault:"false"`
	GRPCPort    int    `env:"OPENMOCK_GRPC_PORT" envDefault:"50051"`
	GRPCHost    string `env:"OPENMOCK_GRPC_HOST" envDefault:"0.0.0.0"`

	//////////////////////// Customized functions //////////////////////////////////
	// KafkaConsumePipelineFunc is a pipeline function run to when consume a message
	KafkaConsumePipelineFunc KafkaPipelineFunc
	// KafkaPublishPipelineFunc is a pipeline function run to when produce a message
	KafkaPublishPipelineFunc KafkaPipelineFunc
	// GRPCServiceMap is a map of gRPC [service_name => GRPCService]
	GRPCServiceMap map[string]GRPCService
	////////////////////////////////////////////////////////////////////////////////

	// Private
	repo        *MockRepo
	kafkaClient *kafkaClient
	redis       RedisDoer
}

func (om *OpenMock) ToYAML() []byte {
	return om.repo.ToYAML()
}

func (om *OpenMock) ToArray() []*Mock {
	return om.repo.AsArray()
}

// ParseEnv loads env vars into the openmock struct
func (om *OpenMock) ParseEnv() {
	err := env.Parse(om)
	if err != nil {
		log.Fatal(err)
	}
}

func (om *OpenMock) SetupLogrus() {
	l, err := logrus.ParseLevel(om.LogLevel)
	if err != nil {
		logrus.WithField("err", err).Fatalf("failed to set logrus level:%s", om.LogLevel)
	}
	logrus.SetLevel(l)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func (om *OpenMock) SetupRepo() {
	om.repo = &MockRepo{
		HTTPMocks:  HTTPMocks{},
		KafkaMocks: KafkaMocks{},
		AMQPMocks:  AMQPMocks{},
		GRPCMocks:  GRPCMocks{},
		Templates:  MocksArray{},
		Behaviors:  orderedmap.NewOrderedMap(),
	}
}

// Start starts the openmock
func (om *OpenMock) Start() {
	om.SetupLogrus()
	om.SetupRepo()
	om.SetRedis()

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
	if om.GRPCEnabled {
		go om.startGRPC()
	}

	if om.TemplatesDirHotReload {
		go func() {
			err := reload.Do(logrus.Infof, reload.Dir(om.TemplatesDir, reload.Exec))
			if err != nil {
				logrus.Fatal(err)
			}
		}()
	}
}

// Stop clean up and release some resources, it's optional.
func (om *OpenMock) Stop() {
	logrus.Info("Stopping openmock...")
	om.kafkaClient.close()
}

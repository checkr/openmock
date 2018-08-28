package main

import (
	"github.com/caarlos0/env"
	"github.com/checkr/openmock"
)

// ENV represents all the env variables
var ENV = struct {
	TemplatesDir string `env:"OPENMOCK_TEMPLATES_DIR" envDefault:"./templates"`

	HTTPEnabled bool   `env:"OPENMOCK_HTTP_ENABLED" envDefault:"true"`
	HTTPPort    int    `env:"OPENMOCK_HTTP_PORT" envDefault:"9999"`
	HTTPHost    string `env:"OPENMOCK_HTTP_HOST" envDefault:"localhost"`

	KafkaEnabled     bool     `env:"OPENMOCK_KAFKA_ENABLED" envDefault:"false"`
	KafkaClientID    string   `env:"OPENMOCK_KAFKA_CLIENT_ID" envDefault:"openmock"`
	KafkaSeedBrokers []string `env:"OPENMOCK_KAFKA_SEED_BROKERS" envDefault:"kafka:9092,localhost:9092" envSeparator:","`

	AMQPEnabled bool   `env:"OPENMOCK_AMQP_ENABLED" envDefault:"false"`
	AMQPURL     string `env:"OPENMOCK_AMQP_URL" envDefault:"amqp://guest:guest@localhost:5672"`

	RedisType string `env:"OPENMOCK_REDIS_TYPE" envDefault:"memory"`
	RedisURL  string `env:"OPENMOCK_REDIS_URL" envDefault:"localhost:6379"`
}{}

func main() {
	env.Parse(&ENV)
	om := openmock.OpenMock{
		TemplatesDir: ENV.TemplatesDir,

		HTTPEnabled: ENV.HTTPEnabled,
		HTTPPort:    ENV.HTTPPort,
		HTTPHost:    ENV.HTTPHost,

		KafkaEnabled:     ENV.KafkaEnabled,
		KafkaClientID:    ENV.KafkaClientID,
		KafkaSeedBrokers: ENV.KafkaSeedBrokers,

		AMQPEnabled: ENV.AMQPEnabled,
		AMQPURL:     ENV.AMQPURL,

		RedisType: ENV.RedisType,
		RedisURL:  ENV.RedisURL,
	}
	om.Start()
}

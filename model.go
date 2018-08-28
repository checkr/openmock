package openmock

import "time"

// Mock represents a mock struct
type Mock struct {
	Key     string   `yaml:"key"`
	Expect  Expect   `yaml:"expect"`
	Actions []Action `yaml:"actions"`
}

type (
	// MocksArray represents an array of Mocks
	MocksArray []*Mock

	// HTTPMocks ...
	HTTPMocks map[ExpectHTTP]MocksArray

	// KafkaMocks is keyed by Topic
	KafkaMocks map[ExpectKafka]MocksArray

	// AMQPMocks is keyed by Queue name
	AMQPMocks map[ExpectAMQP]MocksArray

	// MockRepo stores a repository of Mocks
	MockRepo struct {
		HTTPMocks  HTTPMocks
		KafkaMocks KafkaMocks
		AMQPMocks  AMQPMocks
	}
)

type (
	// Expect represents what to expect from a mock
	Expect struct {
		HTTP      ExpectHTTP  `yaml:"http"`
		Kafka     ExpectKafka `yaml:"kafka"`
		AMQP      ExpectAMQP  `yaml:"amqp"`
		Condition string      `yaml:"condition"`
	}

	// ExpectKafka represents kafka expectation
	ExpectKafka struct {
		Topic string `yaml:"topic"`
	}

	// ExpectAMQP represents amqp expectation
	ExpectAMQP struct {
		Exchange   string `yaml:"exchange"`
		RoutingKey string `yaml:"routing_key"`
		Queue      string `yaml:"queue"`
	}

	// ExpectHTTP represents http expectation
	ExpectHTTP struct {
		Method string `yaml:"method"`
		Path   string `yaml:"path"`
	}
)

// Action represents actions
type Action struct {
	ActionSleep        ActionSleep        `yaml:"sleep"`
	ActionRedis        ActionRedis        `yaml:"redis"`
	ActionPublishKafka ActionPublishKafka `yaml:"publish_kafka"`
	ActionPublishAMQP  ActionPublishAMQP  `yaml:"publish_amqp"`
	ActionSendWebhook  ActionSendWebhook  `yaml:"send_webhook"`
	ActionReplyHTTP    ActionReplyHTTP    `yaml:"reply_http"`
}

// ActionRedis represents a list of redis commands
type ActionRedis []string

// ActionSendWebhook represents the webhook return
type ActionSendWebhook struct {
	URL          string            `yaml:"url"`
	StatusCode   int               `yaml:"status_code"`
	Headers      map[string]string `yaml:"headers"`
	Body         string            `yaml:"body"`
	BodyFromFile string            `yaml:"body_from_file"`
}

// ActionReplyHTTP represents http return
type ActionReplyHTTP struct {
	StatusCode   int               `yaml:"status_code"`
	Headers      map[string]string `yaml:"headers"`
	Body         string            `yaml:"body"`
	BodyFromFile string            `yaml:"body_from_file"`
}

// ActionPublishAMQP represents AMQP return
type ActionPublishAMQP struct {
	Exchange        string `yaml:"exchange"`
	RoutingKey      string `yaml:"routing_key"`
	Payload         string `yaml:"payload"`
	PayloadFromFile string `yaml:"payload_from_file"`
}

// ActionPublishKafka represents kafka return
type ActionPublishKafka struct {
	Topic           string `yaml:"topic"`
	Payload         string `yaml:"payload"`
	PayloadFromFile string `yaml:"payload_from_file"`
}

// ActionSleep represents the sleep action
type ActionSleep struct {
	Duration time.Duration `yaml:"duration"`
}

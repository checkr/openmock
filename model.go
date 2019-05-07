package openmock

import (
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Mock represents a mock struct
type Mock struct {
	Key     string   `yaml:"key,omitempty"`
	Expect  Expect   `yaml:"expect,omitempty"`
	Actions []Action `yaml:"actions,omitempty"`
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
		HTTP      ExpectHTTP  `yaml:"http,omitempty"`
		Kafka     ExpectKafka `yaml:"kafka,omitempty"`
		AMQP      ExpectAMQP  `yaml:"amqp,omitempty"`
		Condition string      `yaml:"condition,omitempty"`
	}

	// ExpectKafka represents kafka expectation
	ExpectKafka struct {
		Topic string `yaml:"topic,omitempty"`
	}

	// ExpectAMQP represents amqp expectation
	ExpectAMQP struct {
		Exchange   string `yaml:"exchange,omitempty"`
		RoutingKey string `yaml:"routing_key,omitempty"`
		Queue      string `yaml:"queue,omitempty"`
	}

	// ExpectHTTP represents http expectation
	ExpectHTTP struct {
		Method string `yaml:"method,omitempty"`
		Path   string `yaml:"path,omitempty"`
	}
)

// Action represents actions
type Action struct {
	ActionPublishAMQP  ActionPublishAMQP  `yaml:"publish_amqp,omitempty"`
	ActionPublishKafka ActionPublishKafka `yaml:"publish_kafka,omitempty"`
	ActionRedis        ActionRedis        `yaml:"redis,omitempty"`
	ActionReplyHTTP    ActionReplyHTTP    `yaml:"reply_http,omitempty"`
	ActionSendHTTP     ActionSendHTTP     `yaml:"send_http,omitempty"`
	ActionSleep        ActionSleep        `yaml:"sleep,omitempty"`
}

// ActionRedis represents a list of redis commands
type ActionRedis []string

// ActionSendHTTP represents the send http action
type ActionSendHTTP struct {
	URL          string            `yaml:"url,omitempty"`
	Method       string            `yaml:"method,omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty"`
	Body         string            `yaml:"body,omitempty"`
	BodyFromFile string            `yaml:"body_from_file,omitempty"`
}

// ActionReplyHTTP represents reply http action
type ActionReplyHTTP struct {
	StatusCode   int               `yaml:"status_code,omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty"`
	Body         string            `yaml:"body,omitempty"`
	BodyFromFile string            `yaml:"body_from_file,omitempty"`
}

// ActionPublishAMQP represents publish AMQP action
type ActionPublishAMQP struct {
	Exchange        string `yaml:"exchange,omitempty"`
	RoutingKey      string `yaml:"routing_key,omitempty"`
	Payload         string `yaml:"payload,omitempty"`
	PayloadFromFile string `yaml:"payload_from_file,omitempty"`
}

// ActionPublishKafka represents publish kafka action
type ActionPublishKafka struct {
	Topic           string `yaml:"topic,omitempty"`
	Payload         string `yaml:"payload,omitempty"`
	PayloadFromFile string `yaml:"payload_from_file,omitempty"`
}

// ActionSleep represents the sleep action
type ActionSleep struct {
	Duration time.Duration `yaml:"duration,omitempty"`
}

// ToYAML outputs MockRepo to yaml bytes
func (repo *MockRepo) ToYAML() []byte {
	ret := []*Mock{}

	for _, arr := range repo.HTTPMocks {
		for _, m := range arr {
			ret = append(ret, m)
		}
	}
	for _, arr := range repo.AMQPMocks {
		for _, m := range arr {
			ret = append(ret, m)
		}
	}
	for _, arr := range repo.KafkaMocks {
		for _, m := range arr {
			ret = append(ret, m)
		}
	}
	b, _ := yaml.Marshal(ret)
	return b
}

package openmock

import (
	"fmt"
	"time"

	"github.com/fatih/structs"
	yaml "gopkg.in/yaml.v2"
)

const (
	// KindBehavior is the default kind
	KindBehavior = "Behavior"

	// KindTemplate is the template kind
	KindTemplate = "Template"
)

// Mock represents a mock struct
type Mock struct {
	// Common fields
	Kind    string `yaml:"kind,omitempty"`
	Key     string `yaml:"key,omitempty"`
	Include string `yaml:"include,omitempty"`

	// KindBehavior fields
	Expect  Expect                 `yaml:"expect,omitempty"`
	Actions []ActionDispatcher     `yaml:"actions,omitempty"`
	Values  map[string]interface{} `yaml:"values,omitempty"`

	// KindTemplate fields
	Template string `yaml:"template,omitempty"`
}

// Validate validates the mock
func (m *Mock) Validate() error {
	if m.Kind == "" {
		m.Kind = KindBehavior
	}

	if m.Key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	switch m.Kind {
	case KindTemplate:
		if !structs.IsZero(m.Expect) || len(m.Actions) != 0 {
			return fmt.Errorf("kind template is only permitted to have `key` and `template` fields. found in: %s", m.Key)
		}
	case KindBehavior:
		if len(m.Template) != 0 {
			return fmt.Errorf("kind behavior is only permitted to have `key`, `expect` and `actions` fields. found in: %s", m.Key)
		}
	default:
		return fmt.Errorf(
			"invalid kind: %s with key: %s. only supported kinds are %v",
			m.Kind, m.Key,
			[]string{KindBehavior, KindTemplate},
		)
	}

	return nil
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
		Templates  MocksArray
		Behaviors  map[string]*Mock
	}
)

type (
	// Expect represents what to expect from a mock
	Expect struct {
		Condition string      `yaml:"condition,omitempty"`
		HTTP      ExpectHTTP  `yaml:"http,omitempty"`
		Kafka     ExpectKafka `yaml:"kafka,omitempty"`
		AMQP      ExpectAMQP  `yaml:"amqp,omitempty"`
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
type ActionDispatcher struct {
	ActionPublishAMQP  ActionPublishAMQP  `yaml:"publish_amqp,omitempty"`
	ActionPublishKafka ActionPublishKafka `yaml:"publish_kafka,omitempty"`
	ActionRedis        ActionRedis        `yaml:"redis,omitempty"`
	ActionReplyHTTP    ActionReplyHTTP    `yaml:"reply_http,omitempty"`
	ActionSendHTTP     ActionSendHTTP     `yaml:"send_http,omitempty"`
	ActionSleep        ActionSleep        `yaml:"sleep,omitempty"`
}

type Action interface {
	Perform(context Context) error
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

	for _, item := range repo.Templates {
		ret = append(ret, item)
	}

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

var getActualAction = func(action ActionDispatcher) Action {
	if !structs.IsZero(action.ActionPublishAMQP) {
		return action.ActionPublishAMQP
	}
	if !structs.IsZero(action.ActionSendHTTP) {
		return action.ActionSendHTTP
	}
	if !structs.IsZero(action.ActionReplyHTTP) {
		return action.ActionReplyHTTP
	}
	if len(action.ActionRedis) > 0 {
		return action.ActionRedis
	}
	if !structs.IsZero(action.ActionPublishKafka) {
		return action.ActionPublishKafka
	}
	return action.ActionSleep
}

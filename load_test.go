package openmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFile(t *testing.T) {
	t.Run("kafka payload", func(t *testing.T) {
		m := &Mock{
			Actions: []ActionDispatcher{
				{
					ActionPublishKafka: ActionPublishKafka{
						PayloadFromFile: "./files/colors.json",
					},
				},
			},
		}
		m.loadFile("demo_templates")
		assert.NotZero(t, m.Actions[0].ActionPublishKafka.Payload)
	})

	t.Run("amqp payload", func(t *testing.T) {
		m := &Mock{
			Actions: []ActionDispatcher{
				{
					ActionPublishAMQP: ActionPublishAMQP{
						PayloadFromFile: "./files/colors.json",
					},
				},
			},
		}
		m.loadFile("demo_templates")
		assert.NotZero(t, m.Actions[0].ActionPublishAMQP.Payload)
	})

	t.Run("http body", func(t *testing.T) {
		m := &Mock{
			Actions: []ActionDispatcher{
				{
					ActionReplyHTTP: ActionReplyHTTP{
						BodyFromFile: "./files/colors.json",
					},
				},
			},
		}
		m.loadFile("demo_templates")
		assert.NotZero(t, m.Actions[0].ActionReplyHTTP.Body)
	})

	t.Run("file not found", func(t *testing.T) {
		m := &Mock{
			Actions: []ActionDispatcher{
				{
					ActionReplyHTTP: ActionReplyHTTP{
						BodyFromFile: "./files/not_exists.json",
					},
				},
			},
		}
		m.loadFile("demo_templates")
		assert.Zero(t, m.Actions[0].ActionReplyHTTP.Body)
	})
}

func TestLoadYAML(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		bytes, err := loadYAML("demo_templates")
		assert.NoError(t, err)
		assert.NotZero(t, bytes)
	})

	t.Run("path not exists", func(t *testing.T) {
		_, err := loadYAML("")
		assert.Error(t, err)
		assert.Equal(t, "lstat : no such file or directory", err.Error())
	})
}

func TestLoad(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		om := &OpenMock{
			TemplatesDir: "demo_templates",
		}
		om.setupRepo()
		err := om.Load()
		assert.NoError(t, err)
		assert.NotZero(t, len(om.repo.AMQPMocks))
		assert.NotZero(t, len(om.repo.HTTPMocks))
		assert.NotZero(t, len(om.repo.KafkaMocks))
	})
}

func TestLoadBehaviors(t *testing.T) {
	om := &OpenMock{}
	om.setupRepo()
	ping := &Mock{
		Key: "ping",
	}
	om.populateBehaviors(MocksArray{ping})
	assert.Equal(t, ping, om.repo.Behaviors["ping"])
}

func TestLoadIncludedBehaviors(t *testing.T) {
	om := &OpenMock{}
	om.setupRepo()
	response := &Mock{
		Key: "response",
		Actions: []ActionDispatcher{{
			ActionReplyHTTP: ActionReplyHTTP{Body: "banana"},
		}},
	}
	ping := &Mock{
		Key:     "ping",
		Include: "response",
		Values: map[string]interface{}{
			"foo": "bar",
		},
	}
	om.populateBehaviors(MocksArray{response, ping})

	t.Run("inherits actions", func(t *testing.T) {
		assert.NotZero(t, om.repo.Behaviors["ping"].Actions)
		assert.Equal(t, "banana", om.repo.Behaviors["ping"].Actions[0].ActionReplyHTTP.Body)
	})

	t.Run("persists values", func(t *testing.T) {
		assert.Equal(t, "bar", om.repo.Behaviors["ping"].Values["foo"])
	})
}

func TestMockPatching(t *testing.T) {
	parentMock := Mock{
		Key: "parent",
		Expect: Expect{
			Condition: "foo",
		},
		Values: map[string]interface{}{
			"value": "not-nana",
		},
	}
	childMock := Mock{
		Key: "child",
		Values: map[string]interface{}{
			"value": "banana",
		},
	}

	patchedMock := parentMock.patchedWith(childMock)

	t.Run("patched mock has values from the parent", func(t *testing.T) {
		assert.Equal(t, "foo", patchedMock.Expect.Condition)
	})

	t.Run("patched mock has values from the patch", func(t *testing.T) {
		assert.Equal(t, "banana", patchedMock.Values["value"])
		assert.Equal(t, "child", patchedMock.Key)
	})

	t.Run("the original mocks are unchanged", func(t *testing.T) {
		assert.Equal(t, "not-nana", parentMock.Values["value"])
		assert.Zero(t, childMock.Expect)
	})
}

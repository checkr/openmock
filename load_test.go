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

func TestLoadExtendedBehaviors(t *testing.T) {
	expectHTTP := ExpectHTTP{
		Method: "GET",
		Path:   "/health-check",
	}

	abstract := &Mock{
		Key:    "abstract",
		Kind:   KindAbstractBehavior,
		Expect: Expect{HTTP: expectHTTP},
		Actions: []ActionDispatcher{{
			ActionReplyHTTP: ActionReplyHTTP{Body: "banana"},
		}},
	}
	concrete := &Mock{
		Key:    "concrete",
		Expect: Expect{HTTP: expectHTTP},
		Extend: "abstract",
		Values: map[string]interface{}{
			"foo": "bar",
		},
	}

	t.Run("inherits actions", func(t *testing.T) {
		om := &OpenMock{}
		om.setupRepo()
		om.populateBehaviors(MocksArray{abstract, concrete})
		assert.NotZero(t, om.repo.Behaviors["concrete"].Actions)
		assert.Equal(t, "banana", om.repo.Behaviors["concrete"].Actions[0].ActionReplyHTTP.Body)
	})

	t.Run("persists values", func(t *testing.T) {
		om := &OpenMock{}
		om.setupRepo()
		om.populateBehaviors(MocksArray{abstract, concrete})
		assert.Equal(t, "bar", om.repo.Behaviors["concrete"].Values["foo"])
	})

	t.Run("extended behavior defined after concrete behavior", func(t *testing.T) {
		om := &OpenMock{}
		om.setupRepo()
		om.populateBehaviors(MocksArray{concrete, abstract})
		assert.Equal(t, "banana", om.repo.Behaviors["concrete"].Actions[0].ActionReplyHTTP.Body)
	})

	t.Run("abstract behavior not exposed as actionable", func(t *testing.T) {
		om := &OpenMock{}
		om.setupRepo()
		om.populateBehaviors(MocksArray{concrete, abstract})
		assert.Equal(t, 1, len(om.repo.HTTPMocks[expectHTTP]))
	})
}

func TestMockPatching(t *testing.T) {
	parentMock := Mock{
		Key: "parent",
		Expect: Expect{
			Condition: "foo",
		},
		Values: map[string]interface{}{
			"foo":   "bar",
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

	t.Run("patched mock overrides values from the patch", func(t *testing.T) {
		assert.Equal(t, "banana", patchedMock.Values["value"])
		assert.Equal(t, "child", patchedMock.Key)
	})

	t.Run("patched mock has values from the parent", func(t *testing.T) {
		assert.Equal(t, "bar", patchedMock.Values["foo"])
	})

	t.Run("the original mocks are unchanged", func(t *testing.T) {
		assert.Equal(t, "not-nana", parentMock.Values["value"])
		assert.Zero(t, childMock.Expect)
	})
}

func TestOverrideBehaviorByKey(t *testing.T) {
	expectHTTP := ExpectHTTP{
		Method: "GET",
		Path:   "/health-check",
	}

	nothealthy := &Mock{
		Key: "health-check",
		Expect: Expect{
			HTTP: expectHTTP,
		},
		Values: map[string]interface{}{
			"value": "not-healthy-banana",
		},
	}
	healthy := &Mock{
		Key: "health-check",
		Expect: Expect{
			HTTP: expectHTTP,
		},
		Values: map[string]interface{}{
			"value": "healthy-banana",
		},
	}

	om := &OpenMock{}
	om.setupRepo()
	om.populateBehaviors(MocksArray{nothealthy, healthy})

	t.Run("should only have one behavior with the same key", func(t *testing.T) {
		assert.Equal(t, 1, len(om.repo.HTTPMocks[expectHTTP]))
	})
	t.Run("should take values with latest behavior", func(t *testing.T) {
		assert.Equal(t, "healthy-banana", om.repo.HTTPMocks[expectHTTP][0].Values["value"])
	})
}

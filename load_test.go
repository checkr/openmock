package openmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFile(t *testing.T) {
	t.Run("kafka payload", func(t *testing.T) {
		m := &Mock{
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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

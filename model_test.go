package openmock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModelValidate(t *testing.T) {

	t.Run("invalid empty", func(t *testing.T) {
		m := &Mock{}
		assert.Error(t, m.Validate())
	})

	t.Run("valid KindTemplate", func(t *testing.T) {
		m := &Mock{
			Kind:     KindTemplate,
			Key:      "t1",
			Template: "t1",
		}
		assert.NoError(t, m.Validate())
	})

	t.Run("invalid KindTemplate - template cannot have non-empty expect", func(t *testing.T) {
		m := &Mock{
			Kind:     KindTemplate,
			Key:      "t1",
			Template: "t1",
			Expect: Expect{
				HTTP: ExpectHTTP{
					Path: "/t1",
				},
			},
		}
		assert.Error(t, m.Validate())
	})

	t.Run("invalid KindTemplate - template cannot have non-empty actions", func(t *testing.T) {
		m := &Mock{
			Kind:     KindTemplate,
			Key:      "t1",
			Template: "t1",
			Actions: []ActionDispatcher{
				{
					ActionSleep: ActionSleep{Duration: time.Second},
				},
			},
		}
		assert.Error(t, m.Validate())
	})

	t.Run("valid KindBehavior", func(t *testing.T) {
		m := &Mock{
			Key: "t1",
			Expect: Expect{
				HTTP: ExpectHTTP{
					Path: "/t1",
				},
			},
			Actions: []ActionDispatcher{
				{
					ActionSleep: ActionSleep{Duration: time.Second},
				},
			},
		}
		assert.NoError(t, m.Validate())
	})

	t.Run("invalid KindBehavior - behavior cannot have template field", func(t *testing.T) {
		m := &Mock{
			Key:      "t1",
			Template: "t1",
			Expect: Expect{
				HTTP: ExpectHTTP{
					Path: "/t1",
				},
			},
			Actions: []ActionDispatcher{
				{
					ActionSleep: ActionSleep{Duration: time.Second},
				},
			},
		}
		assert.Error(t, m.Validate())
	})

	t.Run("invalid KindBehavior - behavior cannot have multiple reply_http actions", func(t *testing.T) {
		m := &Mock{
			Key:      "t1",
			Template: "t1",
			Expect: Expect{
				HTTP: ExpectHTTP{
					Path: "/t1",
				},
			},
			Actions: []ActionDispatcher{
				{
					ActionReplyHTTP: ActionReplyHTTP{StatusCode: 200, Body: "OK"},
				},
				{
					ActionReplyHTTP: ActionReplyHTTP{StatusCode: 200, Body: "GREAT"},
				},
			},
		}
		assert.Error(t, m.Validate())
	})

	t.Run("invalid Kind", func(t *testing.T) {
		m := &Mock{
			Kind: "invalid",
			Key:  "t1",
		}
		assert.Error(t, m.Validate())
	})
}

func TestActionDispatch(t *testing.T) {
	t.Run("resolves to ActionSleep", func(t *testing.T) {
		expectedAction := ActionSleep{Duration: 9001}
		a := ActionDispatcher{
			ActionSleep: expectedAction,
		}
		assert.Equal(t, getActualAction(a), expectedAction)
	})
	t.Run("resolves to ActionSendHTTP", func(t *testing.T) {
		expectedAction := ActionSendHTTP{URL: "potato"}
		a := ActionDispatcher{
			ActionSendHTTP: expectedAction,
		}
		assert.Equal(t, getActualAction(a), expectedAction)
	})
	t.Run("resolves to ActionReplyHTTP", func(t *testing.T) {
		expectedAction := ActionReplyHTTP{StatusCode: 9001}
		a := ActionDispatcher{
			ActionReplyHTTP: expectedAction,
		}
		assert.Equal(t, getActualAction(a), expectedAction)
	})
	t.Run("resolves to ActionReplyGRPC", func(t *testing.T) {
		expectedAction := ActionReplyGRPC{Payload: "success"}
		a := ActionDispatcher{
			ActionReplyGRPC: expectedAction,
		}
		assert.Equal(t, getActualAction(a), expectedAction)
	})
	t.Run("resolves to ActionRedis", func(t *testing.T) {
		expectedAction := ActionRedis{"potato", "potato"}
		a := ActionDispatcher{
			ActionRedis: expectedAction,
		}
		assert.Equal(t, getActualAction(a), expectedAction)
	})
	t.Run("resolves to ActionPublishKafka", func(t *testing.T) {
		expectedAction := ActionPublishKafka{Payload: "potato"}
		a := ActionDispatcher{
			ActionPublishKafka: expectedAction,
		}
		assert.Equal(t, getActualAction(a), expectedAction)
	})
	t.Run("resolves to ActionPublishAMQP", func(t *testing.T) {
		expectedAction := ActionPublishAMQP{Payload: "potato"}
		a := ActionDispatcher{
			ActionPublishAMQP: expectedAction,
		}
		assert.Equal(t, getActualAction(a), expectedAction)
	})
	t.Run("defaults to ActionSleep for duration of 0 (no-op)", func(t *testing.T) {
		expectedAction := ActionSleep{Duration: 0}
		a := ActionDispatcher{}
		assert.Equal(t, getActualAction(a), expectedAction)
	})
}

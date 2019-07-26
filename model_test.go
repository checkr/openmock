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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
				{
					ActionSleep: ActionSleep{Duration: time.Second},
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

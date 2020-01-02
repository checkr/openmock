package openmock

import (
	"testing"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

type FakePerformer struct {
	ReceivedContext Context
	Performed       bool
}

func (fp *FakePerformer) Perform(context Context) {
	fp.Performed = true
	fp.ReceivedContext = context
}

func TestContextUpdate(t *testing.T) {
	t.Run("update context for performing action", func(t *testing.T) {
		mock := Mock{
			Actions: []ActionDispatcher{{}},
			Values: map[string]interface{}{
				"foo": "bar",
			},
		}

		fakePerformer := FakePerformer{}
		defer gostub.StubFunc(&getActualAction, &fakePerformer).Reset()

		err := mock.DoActions(Context{})
		assert.NoError(t, err)
		assert.Equal(t, "bar", fakePerformer.ReceivedContext.Values["foo"])
	})

	t.Run("does not perform action if condition renders falsy", func(t *testing.T) {
		unActingMock := Mock{
			Expect: Expect{
				Condition: "{{.Values.perform}}",
			},
			Actions: []ActionDispatcher{{}},
			Values: map[string]interface{}{
				"perform": false,
			},
		}

		fakePerformer := FakePerformer{}
		defer gostub.StubFunc(&getActualAction, &fakePerformer).Reset()

		err := unActingMock.DoActions(Context{})
		assert.NoError(t, err)
		assert.False(t, fakePerformer.Performed)
	})

	t.Run("does performs action if condition renders truthy", func(t *testing.T) {
		actingMock := Mock{
			Expect: Expect{
				Condition: "{{.Values.perform}}",
			},
			Actions: []ActionDispatcher{{}},
			Values: map[string]interface{}{
				"perform": true,
			},
		}

		fakePerformer := FakePerformer{}
		defer gostub.StubFunc(&getActualAction, &fakePerformer).Reset()

		err := actingMock.DoActions(Context{})
		assert.NoError(t, err)
		assert.True(t, fakePerformer.Performed)
	})
}

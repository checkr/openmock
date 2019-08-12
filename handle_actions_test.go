package openmock

import (
	"testing"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

type FakePerformer struct {
	ReceivedContext Context
}

func (self *FakePerformer) Perform(context Context) error {
	self.ReceivedContext = context
	return nil
}

func TestContextUpdate(t *testing.T) {
	mock := Mock{
		Actions: []ActionDispatcher{ActionDispatcher{}},
		Values: map[string]interface{}{
			"foo": "bar",
		},
	}

	fakePerformer := FakePerformer{}
	defer gostub.StubFunc(&getActualAction, &fakePerformer).Reset()
	mock.DoActions(Context{})

	assert.Equal(t, "bar", fakePerformer.ReceivedContext.Values["foo"])
}

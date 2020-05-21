package evaluator

import (
	"testing"

	models "github.com/checkr/openmock/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestCheckChannelCondition(t *testing.T) {
	context := &models.EvalContext{}

	t.Run("false if context is nil", func(t *testing.T) {
		assert.False(t, checkChannelCondition(nil, nil))
	})

	t.Run("true if http condition is true", func(t *testing.T) {
		defer gostub.StubFunc(&checkHTTPCondition, true).Reset()
		assert.True(t, checkChannelCondition(context, nil))
	})

	t.Run("true if http condition is false but kafka condition is true", func(t *testing.T) {
		defer gostub.StubFunc(&checkHTTPCondition, false).Reset()
		defer gostub.StubFunc(&checkKafkaCondition, true).Reset()
		assert.True(t, checkChannelCondition(context, nil))
	})

	t.Run("false if all conditions are false", func(t *testing.T) {
		defer gostub.StubFunc(&checkHTTPCondition, false).Reset()
		defer gostub.StubFunc(&checkKafkaCondition, false).Reset()
		assert.False(t, checkChannelCondition(context, nil))
	})
}

func TestEvaluate(t *testing.T) {
	t.Run("check channel condition case", func(t *testing.T) {
		t.Run("if channel condition pass", func(t *testing.T) {
			defer gostub.StubFunc(&checkChannelCondition, true).Reset()
			expected_result := models.MockEvalResponse{
				ExpectPassed:     true,
				ActionsPerformed: make([]*models.ActionPerformed, 0, 0),
			}
			actual_result, err := Evaluate(nil, nil)
			assert.Equal(t, expected_result, actual_result)
			assert.Nil(t, err)
		})

		t.Run("if channel condition fail", func(t *testing.T) {
			defer gostub.StubFunc(&checkChannelCondition, false).Reset()
			expected_result := models.MockEvalResponse{
				ExpectPassed:     false,
				ActionsPerformed: make([]*models.ActionPerformed, 0, 0),
			}
			actual_result, err := Evaluate(nil, nil)
			assert.Equal(t, expected_result, actual_result)
			assert.Nil(t, err)
		})
	})
}

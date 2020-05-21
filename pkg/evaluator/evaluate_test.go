package evaluator

import (
	"errors"
	"testing"

	om "github.com/checkr/openmock"
	models "github.com/checkr/openmock/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestCheckCondition(t *testing.T) {
	t.Run("if condition is blank return true", func(t *testing.T) {
		mock := &om.Mock{
			Expect: om.Expect{
				Condition: "",
			},
		}
		passed, rendered, err := checkCondition(nil, mock, nil)
		assert.True(t, passed)
		assert.Equal(t, "", rendered)
		assert.Nil(t, err)
	})

	t.Run("if condition non-blank render it with the conditionContext generated", func(t *testing.T) {
		empty_context := &om.Context{}

		t.Run("with true render returns passed = true", func(t *testing.T) {
			mock := &om.Mock{
				Expect: om.Expect{
					Condition: "true",
				},
			}

			passed, rendered, err := checkCondition(nil, mock, empty_context)
			assert.True(t, passed)
			assert.Equal(t, "true", rendered)
			assert.Nil(t, err)
		})

		t.Run("with false render returns passed = false", func(t *testing.T) {
			mock := &om.Mock{
				Expect: om.Expect{
					Condition: "Foo",
				},
			}

			passed, rendered, err := checkCondition(nil, mock, empty_context)
			assert.False(t, passed)
			assert.Equal(t, "Foo", rendered)
			assert.Nil(t, err)
		})
	})
}

func TestConditionContext(t *testing.T) {
	t.Run("err if context is nil", func(t *testing.T) {
		context, err := conditionContext(nil)
		assert.Nil(t, context)
		assert.NotNil(t, err)
	})

	t.Run("err if all channel contexts are empty", func(t *testing.T) {
		empty_context := &models.EvalContext{
			HTTPContext:  &models.EvalHTTPContext{},
			KafkaContext: &models.EvalKafkaContext{},
		}
		actual_result, err := conditionContext(empty_context)
		assert.Nil(t, actual_result)
		assert.NotNil(t, err)
	})

	t.Run("err if all channel contexts are nil", func(t *testing.T) {
		empty_context := &models.EvalContext{}
		actual_result, err := conditionContext(empty_context)
		assert.Nil(t, actual_result)
		assert.NotNil(t, err)
	})

	t.Run("HTTP Context contains HTTP fields", func(t *testing.T) {
		eval_context := &models.EvalHTTPContext{
			Body: "foobar\nbaz",
			Headers: map[string]string{
				"Header1": "Value1",
				"Header2": "Value2",
			},
			Method:      "GET",
			Path:        "/ping",
			QueryString: "option1=value&option2=value",
		}

		http_context := &models.EvalContext{
			HTTPContext: eval_context,
		}

		defer gostub.StubFunc(&httpToOpenmockConditionContext, nil, nil).Reset()

		actual_result, err := conditionContext(http_context)
		assert.Nil(t, actual_result)
		assert.Nil(t, err)
	})

	t.Run("kafka context contains kafka fields", func(t *testing.T) {
		eval_context := &models.EvalKafkaContext{
			Topic:   "foo",
			Payload: "bar",
		}

		kafka_context := &models.EvalContext{
			KafkaContext: eval_context,
		}

		defer gostub.StubFunc(&kafkaToOpenmockConditionContext, nil, nil).Reset()

		actual_result, err := conditionContext(kafka_context)
		assert.Nil(t, actual_result)
		assert.Nil(t, err)
	})
}

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
	t.Run("happy path", func(t *testing.T) {
		condition_rendered := "did stuff"
		defer gostub.StubFunc(&conditionContext, nil, nil).Reset()
		defer gostub.StubFunc(&checkChannelCondition, true).Reset()
		defer gostub.StubFunc(&checkCondition, true, condition_rendered, nil).Reset()

		expected_result := models.MockEvalResponse{
			ExpectPassed:      true,
			ActionsPerformed:  make([]*models.ActionPerformed, 0, 0),
			ConditionPassed:   true,
			ConditionRendered: condition_rendered,
		}

		actual_result, err := Evaluate(nil, nil)
		assert.Equal(t, expected_result, actual_result)
		assert.Nil(t, err)
	})

	t.Run("if conditionContext err, also err", func(t *testing.T) {

	})

	t.Run("if checkCondition err, also err", func(t *testing.T) {
		expected_err := errors.New("Uhoh")
		defer gostub.StubFunc(&conditionContext, nil, nil).Reset()
		defer gostub.StubFunc(&checkChannelCondition, true).Reset()
		defer gostub.StubFunc(&checkCondition, false, "", expected_err).Reset()

		expected_result := models.MockEvalResponse{
			ExpectPassed:      true,
			ActionsPerformed:  make([]*models.ActionPerformed, 0, 0),
			ConditionPassed:   false,
			ConditionRendered: "",
		}

		actual_result, actual_err := Evaluate(nil, nil)
		assert.Equal(t, expected_result, actual_result)
		assert.Equal(t, expected_err, actual_err)
	})

	t.Run("if checkCondition false, return false", func(t *testing.T) {
		condition_rendered := "foo uhoh"
		defer gostub.StubFunc(&conditionContext, nil, nil).Reset()
		defer gostub.StubFunc(&checkChannelCondition, true).Reset()
		defer gostub.StubFunc(&checkCondition, false, condition_rendered, nil).Reset()

		expected_result := models.MockEvalResponse{
			ExpectPassed:      true,
			ActionsPerformed:  make([]*models.ActionPerformed, 0, 0),
			ConditionPassed:   false,
			ConditionRendered: condition_rendered,
		}

		actual_result, actual_err := Evaluate(nil, nil)
		assert.Equal(t, expected_result, actual_result)
		assert.Nil(t, actual_err)
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
}

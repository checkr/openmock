package evaluator

import (
	"testing"

	om "github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
	"github.com/stretchr/testify/assert"
)

func TestKafkaToOpenmockConditionContext(t *testing.T) {
	t.Run("returns nil when nil input", func(t *testing.T) {
		actual_result, err := kafkaToOpenmockConditionContext(nil)
		assert.Nil(t, actual_result)
		assert.NotNil(t, err)
	})

	t.Run("copies swagger http input context to openmock context", func(t *testing.T) {
		payload := "{\"json\": \"blob\""
		topic := "my_topic"

		eval_context := &models.EvalKafkaContext{
			Payload: payload,
			Topic:   topic,
		}

		expected_result := &om.Context{
			KafkaTopic:   topic,
			KafkaPayload: payload,
		}
		actual_result, err := kafkaToOpenmockConditionContext(eval_context)
		assert.Equal(t, expected_result, actual_result)
		assert.Nil(t, err)
	})
}

func TestCheckKafkaCondition(t *testing.T) {
	matching_topic := "foo"
	mismatching_topic := "bar"

	good_context := &models.EvalKafkaContext{
		Topic: matching_topic,
	}

	good_mock := &om.Mock{
		Expect: om.Expect{
			Kafka: om.ExpectKafka{
				Topic: matching_topic,
			},
		},
	}

	empty_context := &models.EvalKafkaContext{}
	empty_mock := &om.Mock{}

	t.Run("nil mock returns false", func(t *testing.T) {
		assert.False(t, checkKafkaCondition(good_context, nil))
	})

	t.Run("nil context returns false", func(t *testing.T) {
		assert.False(t, checkKafkaCondition(nil, good_mock))
	})

	t.Run("empty mock returns false", func(t *testing.T) {
		assert.False(t, checkKafkaCondition(good_context, empty_mock))
	})

	t.Run("empty context returns false", func(t *testing.T) {
		assert.False(t, checkKafkaCondition(empty_context, good_mock))
	})

	t.Run("topic mismatch returns false", func(t *testing.T) {
		mismatching_context := &models.EvalKafkaContext{
			Topic: mismatching_topic,
		}
		assert.False(t, checkKafkaCondition(mismatching_context, good_mock))
	})

	t.Run("topic match returns true", func(t *testing.T) {
		assert.True(t, checkKafkaCondition(good_context, good_mock))
	})
}

package evaluator

import (
	"errors"

	"github.com/fatih/structs"

	om "github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
)

var kafkaToOpenmockConditionContext = func(context *models.EvalKafkaContext) (*om.Context, error) {
	if context == nil {
		return nil, errors.New("missing input context")
	}

	return &om.Context{
		KafkaTopic:   context.Topic,
		KafkaPayload: context.Payload,
	}, nil
}

var checkKafkaCondition = func(context *models.EvalKafkaContext, mock *om.Mock) bool {
	if context == nil || structs.IsZero(*context) || mock == nil || structs.IsZero(mock.Expect.Kafka) {
		return false
	}

	return context.Topic == mock.Expect.Kafka.Topic
}

package evaluator

import (
	"errors"
	"fmt"

	"github.com/fatih/structs"

	om "github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
)

var performPublishKafkaAction = func(context *om.Context, mock *om.ActionPublishKafka) (*models.PublishKafkaActionPerformed, error) {
	// initial output struct
	out := &models.PublishKafkaActionPerformed{
		Topic: mock.Topic,
	}

	// render payload
	payload, err := context.Render(mock.Payload)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Problem rendering body for publishKafkaAction: %v", err))
	}
	out.Payload = payload

	return out, nil
}

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

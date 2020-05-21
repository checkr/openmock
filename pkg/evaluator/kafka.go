package evaluator

import (
	"github.com/fatih/structs"

	om "github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
)

var checkKafkaCondition = func(context *models.EvalKafkaContext, mock *om.Mock) bool {
	if context == nil || structs.IsZero(*context) || mock == nil || structs.IsZero(mock.Expect.Kafka) {
		return false
	}

	return context.Topic == mock.Expect.Kafka.Topic
}

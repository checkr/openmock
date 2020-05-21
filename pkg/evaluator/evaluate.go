package evaluator

import (
	om "github.com/checkr/openmock"
	models "github.com/checkr/openmock/swagger_gen/models"
)

var Evaluate = func(context *models.EvalContext, mock *om.Mock) (response models.MockEvalResponse, err error) {
	// check if the mock's Expect matches the input context (e.g. HTTP path & method)
	expect_passed := checkChannelCondition(context, mock)
	if !expect_passed {
		return models.MockEvalResponse{
			ExpectPassed:     false,
			ActionsPerformed: make([]*models.ActionPerformed, 0, 0),
		}, nil
	}

	// TODO check if mock's expect condition passes
	// TODO if both match, see what the actions would be

	return models.MockEvalResponse{
		ExpectPassed:     expect_passed,
		ActionsPerformed: make([]*models.ActionPerformed, 0, 0),
	}, nil
}

var checkChannelCondition = func(context *models.EvalContext, mock *om.Mock) bool {
	if context == nil {
		return false
	}
	return checkHTTPCondition(context.HTTPContext, mock) || checkKafkaCondition(context.KafkaContext, mock)
}

package evaluator

import (
	"errors"

	om "github.com/checkr/openmock"
	models "github.com/checkr/openmock/swagger_gen/models"
	"github.com/fatih/structs"
	"github.com/sirupsen/logrus"
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

	// get the OM context that will be used to check condition and perform actions
	om_context, err := conditionContext(context)
	if err != nil {
		logrus.Errorf("Problem setting up om context %v", err)
		return models.MockEvalResponse{
			ExpectPassed:     true,
			ActionsPerformed: make([]*models.ActionPerformed, 0, 0),
		}, err
	}
	if om_context != nil {
		om_context.Values = mock.Values
	}

	// TODO should we set om_context.currentMock? need to make it not private in OM model if so

	// check if mock's expect condition passes
	condition_passed, condition_rendered, err := checkCondition(context, mock, om_context)
	if err != nil {
		return models.MockEvalResponse{
			ExpectPassed:     true,
			ActionsPerformed: make([]*models.ActionPerformed, 0, 0),
		}, err
	}
	if !condition_passed {
		return models.MockEvalResponse{
			ExpectPassed:      true,
			ActionsPerformed:  make([]*models.ActionPerformed, 0, 0),
			ConditionRendered: condition_rendered,
		}, nil
	}

	// TODO if both match, see what the actions would be

	return models.MockEvalResponse{
		ExpectPassed:      expect_passed,
		ActionsPerformed:  make([]*models.ActionPerformed, 0, 0),
		ConditionPassed:   condition_passed,
		ConditionRendered: condition_rendered,
	}, nil
}

var checkChannelCondition = func(context *models.EvalContext, mock *om.Mock) bool {
	if context == nil {
		return false
	}
	return checkHTTPCondition(context.HTTPContext, mock) || checkKafkaCondition(context.KafkaContext, mock)
}

var checkCondition = func(context *models.EvalContext, mock *om.Mock, om_context *om.Context) (bool, string, error) {
	// blank condition is considered a match
	if mock.Expect.Condition == "" {
		return true, "", nil
	}

	// check if condition matches
	render_result, err := om_context.Render(mock.Expect.Condition)
	if err != nil {
		return false, render_result, err
	}

	return render_result == "true", render_result, nil
}

var conditionContext = func(context *models.EvalContext) (*om.Context, error) {
	if context == nil || structs.IsZero(*context) {
		return nil, errors.New("can't make context for nil input")
	}

	if context.HTTPContext != nil && !structs.IsZero(context.HTTPContext) {
		return httpToOpenmockConditionContext(context.HTTPContext)
	}

	if context.KafkaContext != nil && !structs.IsZero(context.KafkaContext) {
		return kafkaToOpenmockConditionContext(context.KafkaContext)
	}

	// TODO - maybe we'd want to do something where we combined the contexts?
	// you could write a mock behavior that responded to either a HTTP call or
	// kafka with the same actions.

	return nil, errors.New("All channels had no context to make condition context")
}

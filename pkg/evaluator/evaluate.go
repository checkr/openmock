package evaluator

import (
	"errors"
	"fmt"
	"sort"

	om "github.com/checkr/openmock"
	models "github.com/checkr/openmock/swagger_gen/models"
	"github.com/fatih/structs"
	"github.com/sirupsen/logrus"
)

var Evaluate = func(context *models.EvalContext, mock *om.Mock) (response models.MockEvalResponse, err error) {
	actions_performed := &[]*models.ActionPerformed{}

	// check if the mock's Expect matches the input context (e.g. HTTP path & method)
	expect_passed := checkChannelCondition(context, mock)
	if !expect_passed {
		return models.MockEvalResponse{
			ExpectPassed:     false,
			ActionsPerformed: *actions_performed,
		}, nil
	}

	// get the OM context that will be used to check condition and perform actions
	om_context, err := conditionContext(context)
	if err != nil {
		logrus.Errorf("Problem setting up om context %v", err)
		return models.MockEvalResponse{
			ExpectPassed:     true,
			ActionsPerformed: *actions_performed,
		}, err
	}
	if om_context != nil {
		om_context.Values = mock.Values
	}

	// check if mock's expect condition passes
	condition_passed, condition_rendered, err := checkCondition(context, mock, om_context)
	if err != nil {
		return models.MockEvalResponse{
			ExpectPassed:     true,
			ActionsPerformed: *actions_performed,
		}, err
	}
	if !condition_passed {
		return models.MockEvalResponse{
			ExpectPassed:      true,
			ActionsPerformed:  *actions_performed,
			ConditionRendered: condition_rendered,
		}, nil
	}

	// if both match, see what the actions would be
	if mock != nil {
		actions_performed, err = actionsPerformed(om_context, &mock.Actions)
		if err != nil {
			return models.MockEvalResponse{
				ExpectPassed:      true,
				ActionsPerformed:  *actions_performed,
				ConditionRendered: condition_rendered,
			}, err
		}
	}

	return models.MockEvalResponse{
		ExpectPassed:      expect_passed,
		ActionsPerformed:  *actions_performed,
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

	out := om.Context{}

	if context.HTTPContext != nil && !structs.IsZero(context.HTTPContext) {
		if http_context, err := httpToOpenmockConditionContext(context.HTTPContext); err == nil {
			out = out.Merge(*http_context)
		} else {
			return &out, fmt.Errorf("Problem constructing http context %v", err)
		}
	}

	if context.KafkaContext != nil && !structs.IsZero(context.KafkaContext) {
		if kafka_context, err := kafkaToOpenmockConditionContext(context.KafkaContext); err == nil {
			out = out.Merge(*kafka_context)
		} else {
			return &out, fmt.Errorf("Problem constructing kafka context %v", err)
		}
	}

	return &out, nil
}

var actionsPerformed = func(context *om.Context, actions *[]om.ActionDispatcher) (*[]*models.ActionPerformed, error) {
	// do actions in specified Order
	ordered_actions := *actions
	sort.Slice(ordered_actions, func(i, j int) bool {
		return ordered_actions[i].Order < ordered_actions[j].Order
	})

	// for each action map it to an output action
	output_actions := make([]*models.ActionPerformed, 0, len(ordered_actions))
	for _, om_action := range ordered_actions {
		actual_action := om.GetActualAction(om_action)

		if reply_http_action, ok := actual_action.(om.ActionReplyHTTP); ok {
			if output_action, err := performReplyHTTPAction(context, &reply_http_action); err == nil {
				output_actions = append(output_actions, &models.ActionPerformed{
					ReplyHTTPActionPerformed: output_action,
				})
			} else {
				return &output_actions, errors.New(fmt.Sprintf("Error performing reply http action %v", err))
			}
		}

		if publish_kafka_action, ok := actual_action.(om.ActionPublishKafka); ok {
			if output_action, err := performPublishKafkaAction(context, &publish_kafka_action); err == nil {
				output_actions = append(output_actions, &models.ActionPerformed{
					PublishKafkaActionPerformed: output_action,
				})
			} else {
				return &output_actions, errors.New(fmt.Sprintf("Error performing publish kafka action %v", err))
			}
		}
	}

	return &output_actions, nil
}

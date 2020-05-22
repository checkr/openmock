package evaluator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	om "github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
)

func TestPerformReplyHTTPAction(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mock_action := &om.ActionReplyHTTP{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: "{\"greeting\": \"hello\", \"path\": \"{{ .HTTPPath }}\"}",
		}

		eval_context := &om.Context{
			HTTPPath: "/ping",
		}

		expected_result := &models.ReplyHTTPActionPerformed{
			Body:        "{\"greeting\": \"hello\", \"path\": \"/ping\"}",
			ContentType: "application/json",
			Headers: map[string]string{
				"Content-Type":   "application/json",
				"Content-Length": "38",
			},
			StatusCode: "200",
		}

		actual_result, err := performReplyHTTPAction(eval_context, mock_action)
		assert.Equal(t, expected_result, actual_result)
		assert.Nil(t, err)
	})
}

func TestHTTPToOpenmockConditionContext(t *testing.T) {
	t.Run("returns nil when nil input", func(t *testing.T) {
		actual_result, err := httpToOpenmockConditionContext(nil)
		assert.Nil(t, actual_result)
		assert.NotNil(t, err)
	})

	t.Run("copies swagger http input context to openmock context", func(t *testing.T) {
		body := "foobar\nbaz"
		method := "GET"
		path := "/ping"
		query_string := "option1=value&option2=value"

		eval_context := &models.EvalHTTPContext{
			Body: body,
			Headers: map[string]interface{}{
				"Header1": "Value1",
				"Header2": "Value2",
			},
			Method:      method,
			Path:        path,
			QueryString: query_string,
		}

		expected_result := &om.Context{
			HTTPBody:        body,
			HTTPPath:        path,
			HTTPQueryString: query_string,
			HTTPHeader: map[string][]string{
				"Header1": []string{"Value1"},
				"Header2": []string{"Value2"},
			},
		}
		actual_result, err := httpToOpenmockConditionContext(eval_context)
		assert.Equal(t, expected_result, actual_result)
		assert.Nil(t, err)
	})
}

func TestCheckHTTPCondition(t *testing.T) {
	matching_method := "GET"
	mismatching_method := "POST"
	matching_path := "/ping"
	mismatching_path := "/pong"

	good_context := &models.EvalHTTPContext{
		Method: matching_method,
		Path:   matching_path,
	}

	empty_context := &models.EvalHTTPContext{}

	good_mock := &om.Mock{
		Expect: om.Expect{
			HTTP: om.ExpectHTTP{
				Method: matching_method,
				Path:   matching_path,
			},
		},
	}

	empty_mock := &om.Mock{}

	t.Run("nil mock returns false", func(t *testing.T) {
		assert.False(t, checkHTTPCondition(good_context, nil))
	})

	t.Run("empty mock returns false", func(t *testing.T) {
		assert.False(t, checkHTTPCondition(good_context, empty_mock))
	})

	t.Run("nil context returns false", func(t *testing.T) {
		assert.False(t, checkHTTPCondition(nil, good_mock))
	})

	t.Run("empty context returns false", func(t *testing.T) {
		assert.False(t, checkHTTPCondition(empty_context, good_mock))
	})

	t.Run("context / mock method mismatch returns false", func(t *testing.T) {
		mismatching_method_context := &models.EvalHTTPContext{
			Method: mismatching_method,
			Path:   matching_path,
		}

		assert.False(t, checkHTTPCondition(mismatching_method_context, good_mock))
	})

	t.Run("context / mock path mismatch returns false", func(t *testing.T) {
		mismatching_path_context := &models.EvalHTTPContext{
			Method: matching_method,
			Path:   mismatching_path,
		}

		assert.False(t, checkHTTPCondition(mismatching_path_context, good_mock))
	})

	t.Run("context / mock method and patch match returns true", func(t *testing.T) {
		assert.True(t, checkHTTPCondition(good_context, good_mock))
	})

	t.Run("with a complicated path", func(t *testing.T) {
		complicated_mock := &om.Mock{
			Expect: om.Expect{
				HTTP: om.ExpectHTTP{
					Method: matching_method,
					Path:   "/path/:id/thing",
				},
			},
		}

		t.Run("non-matching path returns false", func(t *testing.T) {
			non_matching_context := &models.EvalHTTPContext{
				Method: matching_method,
				Path:   "/path/bob",
			}
			assert.False(t, checkHTTPCondition(non_matching_context, complicated_mock))
		})

		t.Run("matching path returns true", func(t *testing.T) {
			matching_context := &models.EvalHTTPContext{
				Method: matching_method,
				Path:   "/path/bob/thing",
			}
			assert.True(t, checkHTTPCondition(matching_context, complicated_mock))
		})
	})
}

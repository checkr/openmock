package evaluator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	om "github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
)

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

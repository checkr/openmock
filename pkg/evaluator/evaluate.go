package evaluator

import (
	"errors"

	om "github.com/checkr/openmock"
	models "github.com/checkr/openmock/swagger_gen/models"
)

type Evaluator interface {
	Evaluate(*models.EvalContext, *om.Mock) (models.MockEvalResponse, error)
}

type evaluator struct {
}

func NewEvaluator() Evaluator {
	return &evaluator{}
}

func (e *evaluator) Evaluate(context *models.EvalContext, mock *om.Mock) (models.MockEvalResponse, error) {
	return models.MockEvalResponse{}, errors.New("Unimplemented")
}

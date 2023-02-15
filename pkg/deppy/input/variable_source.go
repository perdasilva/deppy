package input

import (
	"context"

	"github.com/operator-framework/deppy/pkg/deppy"
)

type VariableSource interface {
	GetVariables(ctx context.Context) ([]deppy.Variable, error)
}

var _ deppy.Variable = &Variable{}

type Variable struct {
	id          deppy.Identifier
	constraints []deppy.Constraint
}

func (s *Variable) Identifier() deppy.Identifier {
	return s.id
}

func (s *Variable) Constraints() []deppy.Constraint {
	return s.constraints
}

func (s *Variable) AddConstraint(constraint deppy.Constraint) {
	s.constraints = append(s.constraints, constraint)
}

func NewVariable(id deppy.Identifier, constraints ...deppy.Constraint) *Variable {
	return &Variable{
		id:          id,
		constraints: constraints,
	}
}

package v2

import (
	"context"

	"github.com/operator-framework/deppy/pkg/sat"
)

// DeppySolver is a simple solver implementation that takes an entity source group and a constraint aggregator
// to produce a Solution (or error if no solution can be found)
type DeppySolver[E Entity, V sat.Variable, Q EntitySource[E]] struct {
	entitySource   Q
	variableSource VariableSource[E, V, Q]
}

func NewDeppySolver[E Entity, V sat.Variable, Q EntitySource[E]](entitySource Q, variableSource VariableSource[E, V, Q]) (*DeppySolver[E, V, Q], error) {
	return &DeppySolver[E, V, Q]{
		entitySource:   entitySource,
		variableSource: variableSource,
	}, nil
}

func (d *DeppySolver[E, V, Q]) Solve(ctx context.Context) (Solution, error) {
	vars, err := d.variableSource.GetVariables(ctx, d.entitySource)
	if err != nil {
		return nil, err
	}

	satSolver, err := sat.NewSolver(sat.WithGenericIntput(vars))
	if err != nil {
		return nil, err
	}

	return satSolver.Solve(ctx)
}

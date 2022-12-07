package v2

import (
	"context"

	"github.com/operator-framework/deppy/pkg/sat"
)

type EntityID string
type EntitySourceID string

type Entity interface {
	ID() EntityID
}

type EntitySource[E Entity] interface {
	ID() EntitySourceID
	Get(context.Context, EntityID) (E, error)
}

type VariableSource[E Entity, V sat.Variable, Q EntitySource[E]] interface {
	GetVariables(ctx context.Context, source Q) ([]V, error)
}

type Solution []sat.Variable

type Solver[V sat.Variable] interface {
	Solve(ctx context.Context) (Solution, error)
}

//type VariableID string
//type EntityID string
//type EntitySourceID string
//
//type Constraint interface {
//	String(subject VariableID) string
//}
//
//type Variable interface {
//	VariableID() VariableID
//	Constraints() []Constraint
//}
//
//type Entity interface {
//	EntityID() EntityID
//}
//
//type EntitySource[E Entity] interface {
//	EntitySourceID() EntitySourceID
//	Get(ctx context.Context, entityID EntityID) (*E, error)
//}
//
//type VariableSource[E Entity, V Variable, Q EntitySource[E]] interface {
//	GetVariables(ctx context.Context, source *Q) ([]V, error)
//}
//
//type Solution[V Variable] []V
//
//type Solver[V Variable] interface {
//	Solve(ctx context.Context, variables []V) (Solution[V], error)
//}
//
//var _ Constraint = &constraint{}
//
//type constraint struct {
//}
//
//func (c constraint) String(subject VariableID) string {
//	//TODO implement me
//	panic("implement me")
//}
//
//var _ Variable = &variable{}
//
//type variable struct {
//}
//
//func (v variable) VariableID() VariableID {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (v variable) Constraints() []Constraint {
//	//TODO implement me
//	panic("implement me")
//}
//
//var _ Entity = &entity{}
//
//type entity struct {
//}
//
//func (e entity) EntityID() EntityID {
//	//TODO implement me
//	panic("implement me")
//}
//
//var _ EntitySource[entity] = &entitysource{}
//
//type entitysource struct {
//}
//
//func (e entitysource) Get(ctx context.Context, entityID EntityID) (*entity, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (e entitysource) EntitySourceID() EntitySourceID {
//	panic("implement me")
//}
//
//var _ VariableSource[entity, variable, entitysource] = &variablesource{}
//
//type variablesource struct {
//}
//
//func (v variablesource) GetVariables(ctx context.Context, source entitysource) ([]variable, error) {
//	//TODO implement me
//	panic("implement me")
//}

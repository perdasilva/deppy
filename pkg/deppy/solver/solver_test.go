package solver_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/operator-framework/deppy/pkg/deppy/input"

	"github.com/operator-framework/deppy/pkg/deppy/solver"

	"github.com/operator-framework/deppy/pkg/deppy/constraint"

	"github.com/operator-framework/deppy/pkg/deppy"
)

func TestSolver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Solver Suite")
}

var _ input.VariableSource = &TestVariableSource{}

type TestVariableSource struct {
	variables []deppy.Variable
}

func (c TestVariableSource) GetVariables(_ context.Context) ([]deppy.Variable, error) {
	return c.variables, nil
}

func NewTestVariableSource(variables ...deppy.Variable) *TestVariableSource {
	return &TestVariableSource{
		variables: variables,
	}
}

var _ = Describe("Entity", func() {
	It("should select a mandatory entity", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Mandatory()),
			input.NewVariable("2"),
		}
		s := NewTestVariableSource(variables...)
		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.SelectedVariables()).To(MatchAllKeys(Keys{
			deppy.Identifier("1"): Equal(input.NewVariable("1", constraint.Mandatory())),
		}))
		Expect(solution.AllVariables()).To(BeNil())
	})

	It("should select two mandatory entities", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Mandatory()),
			input.NewVariable("2", constraint.Mandatory()),
		}
		s := NewTestVariableSource(variables...)
		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.SelectedVariables()).To(MatchAllKeys(Keys{
			deppy.Identifier("1"): Equal(input.NewVariable("1", constraint.Mandatory())),
			deppy.Identifier("2"): Equal(input.NewVariable("2", constraint.Mandatory())),
		}))
	})

	It("should select a mandatory entity and its dependency", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Mandatory(), constraint.Dependency("2")),
			input.NewVariable("2"),
			input.NewVariable("3"),
		}
		s := NewTestVariableSource(variables...)

		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.SelectedVariables()).To(MatchAllKeys(Keys{
			deppy.Identifier("1"): Equal(input.NewVariable("1", constraint.Mandatory(), constraint.Dependency("2"))),
			deppy.Identifier("2"): Equal(input.NewVariable("2")),
		}))
	})

	It("should place resolution errors in the solution", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Mandatory(), constraint.Dependency("2")),
			input.NewVariable("2", constraint.Prohibited()),
			input.NewVariable("3"),
		}
		s := NewTestVariableSource(variables...)
		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.Error()).To(HaveOccurred())
	})

	It("should return peripheral errors", func() {
		so, err := solver.NewDeppySolver(FailingVariableSource{})
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).To(HaveOccurred())
		Expect(solution).To(BeNil())
	})

	It("should select a mandatory entity and its dependency and ignore a non-mandatory prohibited variable", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Mandatory(), constraint.Dependency("2")),
			input.NewVariable("2"),
			input.NewVariable("3", constraint.Prohibited()),
		}
		s := NewTestVariableSource(variables...)
		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.SelectedVariables()).To(MatchAllKeys(Keys{
			deppy.Identifier("1"): Equal(input.NewVariable("1", constraint.Mandatory(), constraint.Dependency("2"))),
			deppy.Identifier("2"): Equal(input.NewVariable("2")),
		}))
	})

	It("should add all variables to solution if option is given", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Mandatory(), constraint.Dependency("2")),
			input.NewVariable("2"),
			input.NewVariable("3", constraint.Prohibited()),
		}
		s := NewTestVariableSource(variables...)
		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background(), solver.AddAllVariablesToSolution())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.AllVariables()).To(Equal([]deppy.Variable{
			input.NewVariable("1", constraint.Mandatory(), constraint.Dependency("2")),
			input.NewVariable("2"),
			input.NewVariable("3", constraint.Prohibited()),
		}))
	})

	It("should not select 'or' paths that are prohibited", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Or("2", false, false), constraint.Dependency("3")),
			input.NewVariable("2", constraint.Dependency("4")),
			input.NewVariable("3", constraint.Prohibited()),
			input.NewVariable("4"),
		}
		s := NewTestVariableSource(variables...)
		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.SelectedVariables()).To(MatchAllKeys(Keys{
			deppy.Identifier("2"): Equal(input.NewVariable("2", constraint.Dependency("4"))),
			deppy.Identifier("4"): Equal(input.NewVariable("4")),
		}))
	})

	It("should respect atMost constraint", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Or("2", false, false), constraint.Dependency("3"), constraint.Dependency("4")),
			input.NewVariable("2", constraint.Dependency("3")),
			input.NewVariable("3", constraint.AtMost(1, "3", "4")),
			input.NewVariable("4"),
		}
		s := NewTestVariableSource(variables...)
		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.SelectedVariables()).To(MatchAllKeys(Keys{
			deppy.Identifier("2"): Equal(input.NewVariable("2", constraint.Dependency("3"))),
			deppy.Identifier("3"): Equal(input.NewVariable("3", constraint.AtMost(1, "3", "4"))),
		}))
	})

	It("should respect dependency conflicts", func() {
		variables := []deppy.Variable{
			input.NewVariable("1", constraint.Or("2", false, false), constraint.Dependency("3"), constraint.Dependency("4")),
			input.NewVariable("2", constraint.Dependency("4"), constraint.Dependency("5")),
			input.NewVariable("3", constraint.Conflict("6")),
			input.NewVariable("4", constraint.Dependency("6")),
			input.NewVariable("5"),
			input.NewVariable("6"),
		}
		s := NewTestVariableSource(variables...)
		so, err := solver.NewDeppySolver(s)
		Expect(err).ToNot(HaveOccurred())
		solution, err := so.Solve(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(solution.SelectedVariables()).To(MatchAllKeys(Keys{
			deppy.Identifier("2"): Equal(input.NewVariable("2", constraint.Dependency("4"), constraint.Dependency("5"))),
			deppy.Identifier("4"): Equal(input.NewVariable("4", constraint.Dependency("6"))),
			deppy.Identifier("5"): Equal(input.NewVariable("5")),
			deppy.Identifier("6"): Equal(input.NewVariable("6")),
		}))
	})
})

var _ input.VariableSource = &FailingVariableSource{}

type FailingVariableSource struct {
}

func (f FailingVariableSource) GetVariables(_ context.Context) ([]deppy.Variable, error) {
	return nil, fmt.Errorf("error")
}

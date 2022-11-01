package olm

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/operator-framework/deppy/internal/entitysource"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Constraints Suite")
}

type MockQuerier struct{}

func (t MockQuerier) Get(ctx context.Context, id entitysource.EntityID) *entitysource.Entity {
	return &entitysource.Entity{}
}
func (t MockQuerier) Filter(ctx context.Context, filter entitysource.Predicate) (entitysource.EntityList, error) {
	return entitysource.EntityList{}, nil
}
func (t MockQuerier) GroupBy(ctx context.Context, id entitysource.GroupByFunction) (entitysource.EntityListMap, error) {
	return entitysource.EntityListMap{}, nil
}
func (t MockQuerier) Iterate(ctx context.Context, id entitysource.IteratorFunction) error {
	return nil
}

var _ = Describe("Constraints", func() {
	Context("requirePackage", func() {
		Describe("GetVariables", func() {
			var (
				ctx         context.Context
				reqPkg      requirePackage
				mockQuerier MockQuerier
			)
			BeforeEach(func() {
				ctx = context.Background()
				reqPkg = requirePackage{
					// TODO get 'actual' valid format for these
					packageName:  "cool-package-1",
					versionRange: "1.0-2.0",
					channel:      "my-channel",
				}
			})
			It("returns a slice of sat.Variable without errors", func() {
				vars, err := reqPkg.GetVariables(ctx, mockQuerier)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).ShouldNot(BeEmpty())
			})
		})
		Describe("RequirePackage", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
	})
	Context("uniqueness", func() {
		Describe("GetVariables", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("GVKUniqueness", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("PackageUniqueness", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("uniquenessSubjectFormat", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
	})
	Context("packageDependency", func() {
		Describe("GetVariables", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("PackageDependency", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
	})
	Context("gvkDependency", func() {
		Describe("GetVariables", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("GVKDependency", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
	})
	Context("entitySource Predicates", func() {
		Describe("withPackageName", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("withinVersion", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("withChannel", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("withExportsGVK", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
	})
	Context("entity helper funcs", func() {
		Describe("byChannelAndVersion", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("gvkGroupFunction", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("packageGroupFunction", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("subject", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
		Describe("toSatIdentifier", func() {
			var ()
			BeforeEach(func() {

			})
			It("does stuff", func() {
			})
		})
	})
})

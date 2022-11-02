package olm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/operator-framework/deppy/internal/entitysource"
	"github.com/operator-framework/deppy/internal/sat"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Constraints Suite")
}

// MockQuerier type to mock the entity querier
type MockQuerier struct {
	returnError         bool
	errorString         string
	entityListReturn    entitysource.EntityList
	entityListMapReturn entitysource.EntityListMap
}

func (t MockQuerier) Get(ctx context.Context, id entitysource.EntityID) *entitysource.Entity {
	return &entitysource.Entity{}
}
func (t MockQuerier) Filter(ctx context.Context, filter entitysource.Predicate) (entitysource.EntityList, error) {
	if t.returnError {
		return nil, errors.New(t.errorString)
	}
	return t.entityListReturn, nil
}
func (t MockQuerier) GroupBy(ctx context.Context, id entitysource.GroupByFunction) (entitysource.EntityListMap, error) {
	if t.returnError {
		return nil, errors.New(t.errorString)
	}
	return t.entityListMapReturn, nil
}
func (t MockQuerier) Iterate(ctx context.Context, id entitysource.IteratorFunction) error {
	if t.returnError {
		return errors.New(t.errorString)
	}
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
					packageName:  "cool-package-1",
					versionRange: "1.0-2.0",
					channel:      "my-channel",
				}
				mockQuerier = MockQuerier{
					returnError: false,
				}
			})
			It("returns a slice of sat.Variable containing requirement on entity", func() {
				mockQuerier.entityListReturn = entitysource.EntityList{
					*entitysource.NewEntity("cool-package-2-entity", map[string]string{
						propertyOLMPackageName: "cool-package-2",
					}),
				}
				vars, err := reqPkg.GetVariables(ctx, mockQuerier)
				expectedIdentifier := fmt.Sprintf("require-%s-%s-%s", reqPkg.packageName, reqPkg.versionRange, reqPkg.channel)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).Should(HaveLen(1))
				Expect(vars[0]).NotTo(BeNil())
				Expect(vars[0].Identifier().String()).To(Equal(expectedIdentifier))
				Expect(vars[0].Constraints()).Should(HaveLen(2))
				Expect(vars[0].Constraints()[0].String(sat.IdentifierFromString(expectedIdentifier))).To(
					Equal(fmt.Sprintf("%s is mandatory", expectedIdentifier)))
				Expect(vars[0].Constraints()[1].String(sat.IdentifierFromString(expectedIdentifier))).To(
					Equal(fmt.Sprintf("%s requires at least one of cool-package-2-entity", expectedIdentifier)))
			})
			It("forwards any errors from EntityQuerier", func() {
				mockQuerier = MockQuerier{
					returnError: true,
					errorString: "filter failure",
				}
				vars, err := reqPkg.GetVariables(ctx, mockQuerier)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("filter failure"))
				Expect(vars).Should(HaveLen(0))
			})
		})
		Describe("RequirePackage", func() {
			var (
				ctx         context.Context
				mockQuerier MockQuerier
			)
			It("returns constraint generator with values properly set", func() {
				expPackageName := "my-package"
				expVersionRange := "1.1.1-2.2.2"
				expChannel := "my-channel"
				reqPkg := RequirePackage(expPackageName, expVersionRange, expChannel)
				vars, err := reqPkg.GetVariables(ctx, mockQuerier)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).Should(HaveLen(1))
				Expect(vars[0]).NotTo(BeNil())
				Expect(vars[0].Identifier().String()).To(Equal(fmt.Sprintf("require-%s-%s-%s", expPackageName, expVersionRange, expChannel)))
			})
		})
	})
	Context("uniqueness", func() {
		Describe("GetVariables", func() {
			var (
				ctx           context.Context
				mockQuerier   MockQuerier
				entityListMap entitysource.EntityListMap
			)
			BeforeEach(func() {
				ctx = context.Background()
				mockQuerier = MockQuerier{
					returnError: false,
				}
				entityListMap = entitysource.EntityListMap{
					"entity-list-1": entitysource.EntityList{
						*entitysource.NewEntity("entity-1", map[string]string{
							propertyOLMGVK:         "gvk-1",
							propertyOLMPackageName: "package-2",
						}),
					},
					"entity-list-2": entitysource.EntityList{
						*entitysource.NewEntity("entity-2", map[string]string{
							propertyOLMGVK:         "gvk-2",
							propertyOLMPackageName: "package-1",
						}),
					},
				}
				mockQuerier.entityListMapReturn = entityListMap
			})
			// TODO dfranz - how can we test the difference between these two?
			// How many of these vals do we really need to check?
			It("GetVariables with GVKUniqueness", func() {
				vars, err := GVKUniqueness().GetVariables(ctx, mockQuerier)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).Should(HaveLen(len(entityListMap)))
				Expect(vars[0].Identifier().String()).To(Equal("entity-list-1 uniqueness"))
				Expect(vars[0].Constraints()[0].String("entity-1")).To(Equal("entity-1 permits at most 1 of entity-1"))
				Expect(vars[0].Constraints()[0].String("entity-list-1")).To(Equal("entity-list-1 permits at most 1 of entity-1"))
				Expect(vars[1].Identifier().String()).To(Equal("entity-list-2 uniqueness"))
				Expect(vars[1].Constraints()[0].String("entity-2")).To(Equal("entity-2 permits at most 1 of entity-2"))
				Expect(vars[1].Constraints()[0].String("entity-list-2")).To(Equal("entity-list-2 permits at most 1 of entity-2"))
			})
			It("GetVariables with PackageUniqueness", func() {
				vars, err := PackageUniqueness().GetVariables(ctx, mockQuerier)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).Should(HaveLen(len(entityListMap)))
				Expect(vars[0].Identifier().String()).To(Equal("entity-list-1 uniqueness"))
				Expect(vars[0].Constraints()[0].String("entity-1")).To(Equal("entity-1 permits at most 1 of entity-1"))
				Expect(vars[0].Constraints()[0].String("entity-list-1")).To(Equal("entity-list-1 permits at most 1 of entity-1"))
				Expect(vars[1].Identifier().String()).To(Equal("entity-list-2 uniqueness"))
				Expect(vars[1].Constraints()[0].String("entity-2")).To(Equal("entity-2 permits at most 1 of entity-2"))
				Expect(vars[1].Constraints()[0].String("entity-list-2")).To(Equal("entity-list-2 permits at most 1 of entity-2"))
			})
			It("forwards any errors from EntityQuerier", func() {
				mockQuerier = MockQuerier{
					returnError: true,
					errorString: "groupBy failure",
				}
				vars, err := PackageUniqueness().GetVariables(ctx, mockQuerier)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("groupBy failure"))
				Expect(vars).Should(HaveLen(0))
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

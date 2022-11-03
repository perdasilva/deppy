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
			It("returns a slice of sat.Variable containing uniqueness requirements", func() {
				vars, err := GVKUniqueness().GetVariables(ctx, mockQuerier)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).Should(HaveLen(len(entityListMap)))
				// TODO order of these is random
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
		Describe("gvkGroupFunction", func() {
			It("returns a string slice with an entry containing the group, version, and kind delimited by '/'", func() {
				entity := entitysource.NewEntity(
					"my-entity",
					map[string]string{
						propertyOLMGVK: "{\"group\":\"my-group\",\"version\":\"my-version\",\"kind\":\"my-kind\"}",
					},
				)
				ret := gvkGroupFunction(entity)
				Expect(ret).Should(HaveLen(1))
				Expect(ret[0]).To(Equal("my-group/my-version/my-kind"))
			})
			It("returns nothing and does not panic when gvk passed is not parseable as json", func() {
				entity := entitysource.NewEntity(
					"my-entity",
					map[string]string{
						propertyOLMGVK: "what will come back I wonder",
					},
				)
				ret := gvkGroupFunction(entity)
				Expect(ret).Should(HaveLen(0))
			})
			It("returns nothing when 'olm.gvk' key is missing", func() {
				entity := entitysource.NewEntity(
					"my-entity",
					map[string]string{
						"made-up-key": "{\"group\":\"my-group\",\"version\":\"my-version\",\"kind\":\"my-kind\"}",
					},
				)
				ret := gvkGroupFunction(entity)
				Expect(ret).Should(HaveLen(0))
			})
		})
		Describe("packageGroupFunction", func() {
			It("returns a string slice containing a single entry with the package name", func() {
				entity := entitysource.NewEntity(
					"my-entity",
					map[string]string{
						propertyOLMPackageName: "my-cool-package",
					},
				)
				ret := packageGroupFunction(entity)
				Expect(ret).Should(HaveLen(1))
				Expect(ret[0]).To(Equal("my-cool-package"))
			})
			It("returns an empty string slice when 'olm.packageName' key is missing", func() {
				entity := entitysource.NewEntity(
					"my-entity",
					map[string]string{
						"some-key": "my-cool-package",
					},
				)
				ret := packageGroupFunction(entity)
				Expect(ret).Should(HaveLen(0))
			})
		})
	})
	Context("packageDependency", func() {
		Describe("GetVariables", func() {
			var (
				ctx          context.Context
				filterReturn entitysource.EntityList
				pkgDep       packageDependency
				mockQuerier  MockQuerier
			)
			BeforeEach(func() {
				ctx = context.Background()
				pkgDep = packageDependency{
					subject:      "my-subject",
					packageName:  "cool-package-1",
					versionRange: "1.0-2.0",
				}
				filterReturn = entitysource.EntityList{
					*entitysource.NewEntity(
						"entity-1",
						map[string]string{
							"?": "!",
						},
					),
				}
				mockQuerier = MockQuerier{
					returnError:      false,
					entityListReturn: filterReturn,
				}
			})
			It("returns a slice of sat.Variable containing package dependency", func() {
				vars, err := pkgDep.GetVariables(ctx, mockQuerier)
				Expect(err).ToNot(HaveOccurred())
				Expect(vars).To(HaveLen(1))
				Expect(vars[0].Identifier().String()).To(Equal("my-subject"))
				Expect(vars[0].Constraints()[0].String("foo")).To(Equal("foo requires at least one of entity-1"))
			})
			It("forwards any error from the entity querier", func() {
				mockQuerier = MockQuerier{
					returnError: true,
					errorString: "filter failure",
				}
				vars, err := pkgDep.GetVariables(ctx, mockQuerier)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("filter failure"))
				Expect(vars).To(HaveLen(0))
			})
		})
		Describe("PackageDependency", func() {
			var (
				ctx         context.Context
				mockQuerier MockQuerier
			)
			It("constructs a packageDependency object with the provided values", func() {
				expSubject := "my-subject"
				expPackageName := "cool-package-1"
				expVersionRange := "1.1.1-2.2.2"
				pkgDep := PackageDependency(sat.Identifier(expSubject), expPackageName, expVersionRange)

				vars, err := pkgDep.GetVariables(ctx, mockQuerier)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).Should(HaveLen(1))
				Expect(vars[0]).NotTo(BeNil())
				Expect(vars[0].Identifier().String()).To(Equal(expSubject))
			})
		})
	})
	Context("gvkDependency", func() {
		Describe("GetVariables", func() {
			var (
				ctx          context.Context
				mockQuerier  MockQuerier
				filterReturn entitysource.EntityList
				gvkDep       gvkDependency
			)
			BeforeEach(func() {
				ctx = context.Background()
				filterReturn = entitysource.EntityList{
					*entitysource.NewEntity(
						"entity-1",
						map[string]string{
							"foo": "bar",
						},
					),
				}
				gvkDep = gvkDependency{
					subject: "my-subject",
					group:   "my-group",
					version: "3.4",
					kind:    "my-kind",
				}
				mockQuerier = MockQuerier{
					returnError:      false,
					entityListReturn: filterReturn,
				}
			})
			It("returns a slice of sat.Variable containing the gvk dependency", func() {
				vars, err := gvkDep.GetVariables(ctx, mockQuerier)
				Expect(err).ToNot(HaveOccurred())
				Expect(vars).To(HaveLen(1))
				Expect(vars[0].Identifier().String()).To(Equal("my-subject"))
				Expect(vars[0].Constraints()[0].String("package")).To(Equal("package requires at least one of entity-1"))
			})
			It("forwards any error from the entity querier", func() {
				mockQuerier = MockQuerier{
					returnError: true,
					errorString: "filter failure",
				}
				vars, err := gvkDep.GetVariables(ctx, mockQuerier)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("filter failure"))
				Expect(vars).To(HaveLen(0))
			})
		})
		Describe("GVKDependency", func() {
			var (
				ctx         context.Context
				mockQuerier MockQuerier
			)
			It("properly constructs a gvkDependency struct with provided values", func() {
				expSubject := "my-subject"
				expGroup := "my-group"
				expVersion := "1.1.1"
				expKind := "my-kind"
				gvkDep := GVKDependency(sat.Identifier(expSubject), expGroup, expVersion, expKind)
				vars, err := gvkDep.GetVariables(ctx, mockQuerier)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).Should(HaveLen(1))
				Expect(vars[0]).NotTo(BeNil())
				Expect(vars[0].Identifier().String()).To(Equal(expSubject))
			})
		})
	})
	Context("entitySource Predicates", func() {
		DescribeTable("withPackageName", func(packageName string, packageNameKey string, expReturn bool) {
			propertiesMap := map[string]string{
				packageNameKey: "my-package",
			}
			entity := entitysource.NewEntity("e1", propertiesMap)
			pred := withPackageName(packageName)
			result := pred(entity)
			Expect(result).To(Equal(expReturn))
		},
			Entry("gives true result when package name matches value set on 'olm.packageName'", "my-package", propertyOLMPackageName, true),
			Entry("gives false result when package name does not match value set on 'olm.packageName'", "my-other-package", propertyOLMPackageName, false),
			Entry("gives false result when 'olm.packageName' key is missing", "my-package", "irrelevant-key", false),
		)
		DescribeTable("withinVersion", func(entityVersion string, semverRange string, key string, expReturn bool) {
			propertiesMap := map[string]string{
				key: entityVersion,
			}
			entity := entitysource.NewEntity("e1", propertiesMap)
			pred := withinVersion(semverRange)
			result := pred(entity)
			Expect(result).To(Equal(expReturn))
		},
			Entry("gives true result when given a valid version within the specified range", "3.22.1", "<3.22.2", propertyOLMVersion, true),
			Entry("give false result when given a valid version outside the specified range", "3.22.3", "<3.22.2", propertyOLMVersion, false),
			Entry("give false result when given an invalid version", "abcdefg", "<3.22.2", propertyOLMVersion, false),
			Entry("give false result when given an invalid version range", "3.22.1", "abcdefg", propertyOLMVersion, false),
			Entry("give false result when 'olm.version' key is missing from entity", "3.22.1", "<3.22.2", "irrelevent-key", false),
		)
		DescribeTable("withChannel", func(channel string, channelKey string, expReturn bool) {
			propertiesMap := map[string]string{
				channelKey: "a-channel",
			}
			entity := entitysource.NewEntity("e1", propertiesMap)
			pred := withChannel(channel)
			result := pred(entity)
			Expect(result).To(Equal(expReturn))
		},
			Entry("gives true result when provided channel matches value set on 'olm.channel'", "a-channel", propertyOLMChannel, true),
			Entry("gives false result when provided channel does not match value set on 'olm.channel'", "b-channel", propertyOLMChannel, false),
			Entry("gives true result when provided with empty channel", "", propertyOLMChannel, true),
			Entry("gives false result when 'olm.channel' key is missing from entity", "a-channel", "irrelevant-key", false),
		)
		DescribeTable("withExportsGVK", func(group string, gvkKey string, expReturn bool) {
			propertiesMap := map[string]string{
				gvkKey: "{\"group\":\"my-group\",\"version\":\"my-version\",\"kind\":\"my-kind\"}",
			}
			entity := entitysource.NewEntity("e1", propertiesMap)
			pred := withExportsGVK(group, "my-version", "my-kind")
			result := pred(entity)
			Expect(result).To(Equal(expReturn))
		},
			Entry("gives true result when provided group, version, and kind match that in the entity json", "my-group", propertyOLMGVK, true),
			Entry("gives false result when group does not match that in entity json", "my-other-group", propertyOLMGVK, false),
			Entry("gives false result when 'olm.gvk' key is missing from entity", "my-group", "irrelevant-key", false),
		)
	})
	Context("byChannelAndVersion", func() {
		DescribeTable(propertyOLMPackageName, func(e1pkgName string, e2pkgName string, expReturn bool) {
			e1Map := make(map[string]string)
			e2Map := make(map[string]string)
			if e1pkgName != "" {
				e1Map[propertyOLMPackageName] = e1pkgName
			}
			if e2pkgName != "" {
				e2Map[propertyOLMPackageName] = e2pkgName
			}
			e1 := entitysource.NewEntity("e1", e1Map)
			e2 := entitysource.NewEntity("e2", e2Map)
			Expect(byChannelAndVersion(e1, e2)).To(Equal(expReturn))
		},
			Entry("Returns false when e1 package name is missing", "", "", false),
			Entry("Returns true when e2 package name is missing", "p1", "", true),
			Entry("Returns true when e2 package name is >= e1 package name", "p1", "p2", true),
		)
		DescribeTable(propertyOLMChannel, func(e1Channel string, e2Channel string, expReturn bool) {
			e1Map := map[string]string{propertyOLMDefaultChannel: "c2", propertyOLMPackageName: "p"}
			e2Map := map[string]string{propertyOLMDefaultChannel: "c1", propertyOLMPackageName: "p"}
			if e1Channel != "" {
				e1Map[propertyOLMPackageName] = e1Channel
			}
			if e2Channel != "" {
				e2Map[propertyOLMPackageName] = e2Channel
			}
			e1 := entitysource.NewEntity("e1", e1Map)
			e2 := entitysource.NewEntity("e2", e2Map)
			Expect(byChannelAndVersion(e1, e2)).To(Equal(expReturn))
		},
			Entry("Returns false when e1 channel is missing", "", "", false),
			Entry("Returns true when e2 channel is missing", "c1", "", true),
			Entry("Returns e1 channel < e2 channel when e1 and e2 channels do not match their respective defaults", "c1", "c2", true),
		)
		DescribeTable(propertyOLMDefaultChannel, func(e1DefaultChannel string, e2DefaultChannel string, expReturn bool) {
			e1Map := map[string]string{propertyOLMChannel: "c1", propertyOLMPackageName: "p"}
			e2Map := map[string]string{propertyOLMChannel: "c2", propertyOLMPackageName: "p"}
			if e1DefaultChannel != "" {
				e1Map[propertyOLMDefaultChannel] = e1DefaultChannel
			}
			if e2DefaultChannel != "" {
				e2Map[propertyOLMDefaultChannel] = e2DefaultChannel
			}
			e1 := entitysource.NewEntity("e1", e1Map)
			e2 := entitysource.NewEntity("e2", e2Map)
			Expect(byChannelAndVersion(e1, e2)).To(Equal(expReturn))
		},
			Entry("Returns e1channel < e2Channel when e1 default channel is missing", "", "", true),
			Entry("Returns e1channel < e2Channel when e2 default channel is missing", "c1", "", true),
			Entry("Returns true when e1 channel == e1 default channel", "c1", "c2", true),
			Entry("Returns false when e2 channel == e2 default channel", "c2", "c2", false),
		)
		DescribeTable(propertyOLMVersion, func(e1version string, e2version string, expReturn bool) {
			e1Map := map[string]string{propertyOLMChannel: "c1", propertyOLMPackageName: "p"}
			e2Map := map[string]string{propertyOLMChannel: "c1", propertyOLMPackageName: "p"}
			if e1version != "" {
				e1Map[propertyOLMVersion] = e1version
			}
			if e2version != "" {
				e2Map[propertyOLMVersion] = e2version
			}
			e1 := entitysource.NewEntity("e1", e1Map)
			e2 := entitysource.NewEntity("e2", e2Map)
			Expect(byChannelAndVersion(e1, e2)).To(Equal(expReturn))
		},
			Entry("Returns false when e1 version is missing", "", "", false),
			Entry("Returns true when e2 version is missing", "1.0.0", "", true),
			Entry("Returns false when e1 version does not parse", "abcd", "2.0.0", false),
			Entry("Returns false when e2 version does not parse", "2.0.0", "abcd", false),
			Entry("Returns e1 version > e2 version when both are provided and parse correctly", "2.0.0", "1.0.0", true),
		)
	})
	Context("entity helper funcs", func() {
		Describe("subject", func() {
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

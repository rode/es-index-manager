package indexmanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing/fstest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MappingsRegistry", func() {
	var (
		config   *Config
		registry MappingsRegistry
		testFs   fstest.MapFS

		expectedIndexPrefix   string
		expectedMappingDir    string
		expectedDocumentKinds []string
		randomDocumentKind    string
		expectedMapping       *VersionedMapping

		actualLoadMappingsError error
	)

	BeforeEach(func() {
		actualLoadMappingsError = nil
		expectedMappingDir = fake.Word()
		expectedIndexPrefix = fake.Word()

		config = &Config{
			IndexPrefix:  expectedIndexPrefix,
			MappingsPath: expectedMappingDir,
		}

		testFs = fstest.MapFS{}

		expectedDocumentKinds = createDocumentKinds()
		randomDocumentKind = fake.RandomString(expectedDocumentKinds)

		for i := 0; i < len(expectedDocumentKinds); i++ {
			documentKind := expectedDocumentKinds[i]
			mapping := &VersionedMapping{
				Version: fake.Word(),
				Mappings: map[string]interface{}{
					fake.Word(): fake.Word(),
				},
			}

			if documentKind == randomDocumentKind {
				expectedMapping = mapping
			}

			fileName := filepath.Join(expectedMappingDir, fmt.Sprintf("%s.json", documentKind))
			testFs[fileName] = mappingsFile(mapping)
		}
	})

	JustBeforeEach(func() {
		registry = NewMappingsRegistry(config, testFs)

		actualLoadMappingsError = registry.LoadMappings()
	})

	Context("LoadMappings", func() {
		When("mappings are successfully loaded", func() {
			It("should not return an error", func() {
				Expect(actualLoadMappingsError).NotTo(HaveOccurred())
			})
		})

		When("the mappings directory does not exist", func() {
			BeforeEach(func() {
				testFs = fstest.MapFS{}
			})

			It("should return an error", func() {
				Expect(actualLoadMappingsError).To(HaveOccurred())
				Expect(actualLoadMappingsError.Error()).To(ContainSubstring("error finding mappings in directory"))
			})
		})

		When("the file contents are invalid", func() {
			BeforeEach(func() {
				testFs[filepath.Join(expectedMappingDir, randomDocumentKind+".json")].Data = []byte("{")
			})

			It("should return an error", func() {
				Expect(actualLoadMappingsError).To(HaveOccurred())
				Expect(actualLoadMappingsError.Error()).To(ContainSubstring("invalid json"))
			})
		})

		When("there is a subdirectory", func() {
			BeforeEach(func() {
				testFs[filepath.Join(expectedMappingDir, fake.Word())] = &fstest.MapFile{
					Mode: os.ModeDir,
				}
			})

			It("should be ignored", func() {
				Expect(actualLoadMappingsError).To(BeNil())
			})
		})
	})

	Context("IndexName", func() {
		var (
			actualIndexName string
			innerName       string
		)

		BeforeEach(func() {
			innerName = ""
		})

		JustBeforeEach(func() {
			actualIndexName = registry.IndexName(randomDocumentKind, innerName)
		})

		When("an inner name is specified", func() {
			BeforeEach(func() {
				innerName = fake.Word()
			})

			It("should return an index name comprised of the prefix, version, inner name, and document kind", func() {
				expectedIndexName := fmt.Sprintf("%s-%s-%s-%s", expectedIndexPrefix, expectedMapping.Version, innerName, randomDocumentKind)

				Expect(actualIndexName).To(Equal(expectedIndexName))
			})
		})

		When("the inner name contains the delimiter", func() {
			BeforeEach(func() {
				innerName = fmt.Sprintf("%s-%s", fake.Word(), fake.Word())
			})

			It("should return the correct index name", func() {
				expectedIndexName := fmt.Sprintf("%s-%s-%s-%s", expectedIndexPrefix, expectedMapping.Version, innerName, randomDocumentKind)

				Expect(actualIndexName).To(Equal(expectedIndexName))
			})
		})

		When("no inner name is passed", func() {
			It("should return an index name comprised of the prefix, version, and document kind", func() {
				expectedIndexName := fmt.Sprintf("%s-%s-%s", expectedIndexPrefix, expectedMapping.Version, randomDocumentKind)

				Expect(actualIndexName).To(Equal(expectedIndexName))
			})
		})
	})

	Context("AliasName", func() {
		var (
			actualAliasName string
			innerName       string
		)

		BeforeEach(func() {
			innerName = ""
		})

		JustBeforeEach(func() {
			actualAliasName = registry.AliasName(randomDocumentKind, innerName)
		})

		When("an inner name is specified", func() {
			BeforeEach(func() {
				innerName = fake.Word()
			})

			It("should return an alias name comprised of the prefix, inner name, and document kind", func() {
				expectedAliasName := fmt.Sprintf("%s-%s-%s", expectedIndexPrefix, innerName, randomDocumentKind)

				Expect(actualAliasName).To(Equal(expectedAliasName))
			})
		})

		When("no inner name is specified", func() {
			It("should return an alias name comprised of the prefix, and document kind", func() {
				expectedAliasName := fmt.Sprintf("%s-%s", expectedIndexPrefix, randomDocumentKind)

				Expect(actualAliasName).To(Equal(expectedAliasName))
			})
		})
	})

	Context("Version", func() {
		var (
			documentKind  string
			actualVersion string
		)

		BeforeEach(func() {
			documentKind = randomDocumentKind
		})

		JustBeforeEach(func() {
			actualVersion = registry.Version(documentKind)
		})

		When("the document kind is not in the registry", func() {
			BeforeEach(func() {
				documentKind = fake.Word()
			})

			It("should return the empty string", func() {
				Expect(actualVersion).To(BeEmpty())
			})
		})

		When("the document kind is known", func() {
			It("should return the current version", func() {
				Expect(actualVersion).To(Equal(expectedMapping.Version))
			})
		})
	})

	Context("Mapping", func() {
		var (
			documentKind  string
			actualMapping *VersionedMapping
		)

		BeforeEach(func() {
			documentKind = randomDocumentKind
		})

		JustBeforeEach(func() {
			actualMapping = registry.Mapping(documentKind)
		})

		When("the document kind is not in the registry", func() {
			BeforeEach(func() {
				documentKind = fake.Word()
			})

			It("should return nil", func() {
				Expect(actualMapping).To(BeNil())
			})
		})

		When("the document kind is in the registry", func() {
			It("should return the mapping", func() {
				Expect(actualMapping).To(Equal(expectedMapping))
			})
		})
	})
})

func mappingsFile(schema interface{}) *fstest.MapFile {
	data, err := json.Marshal(schema)
	Expect(err).NotTo(HaveOccurred())

	return &fstest.MapFile{
		Data: data,
	}
}

func createDocumentKinds() []string {
	var kinds []string

	for i := 0; i < fake.Number(2, 5); i++ {
		documentKind := fake.Word()
		kinds = append(kinds, documentKind)
	}

	return kinds
}

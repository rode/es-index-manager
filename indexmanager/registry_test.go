// Copyright 2021 The Rode Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indexmanager_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing/fstest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rode/es-index-manager/indexmanager"
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
			mapping := createRandomMapping()

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

	Context("ParseIndexName", func() {
		var (
			indexName       string
			actualIndexName *IndexName
			actualOk        bool
		)

		BeforeEach(func() {
			config.IndexPrefix = "rode"
		})

		JustBeforeEach(func() {
			actualIndexName, actualOk = registry.ParseIndexName(indexName)
		})

		When("the index does not have inner name", func() {
			BeforeEach(func() {
				indexName = "rode-v1alpha1-policies"

				testFs[filepath.Join(config.MappingsPath, "policies.json")] = mappingsFile(&VersionedMapping{
					Version:  fake.Word(),
					Mappings: map[string]interface{}{},
				})
			})

			It("should be parsed correctly", func() {
				Expect(actualIndexName).NotTo(BeNil())
				Expect(actualIndexName.Version).To(Equal("v1alpha1"))
				Expect(actualIndexName.DocumentKind).To(Equal("policies"))
				Expect(actualIndexName.Inner).To(BeEmpty())
				Expect(actualOk).To(BeTrue())
			})
		})

		When("the index has an inner name", func() {
			BeforeEach(func() {
				indexName = "rode-v1alpha1-test-policies"

				testFs[filepath.Join(config.MappingsPath, "policies.json")] = mappingsFile(&VersionedMapping{
					Version:  fake.Word(),
					Mappings: map[string]interface{}{},
				})
			})

			It("should be parsed correctly", func() {
				Expect(actualIndexName).NotTo(BeNil())
				Expect(actualIndexName.Version).To(Equal("v1alpha1"))
				Expect(actualIndexName.DocumentKind).To(Equal("policies"))
				Expect(actualIndexName.Inner).To(Equal("test"))
				Expect(actualOk).To(BeTrue())
			})
		})

		When("the document kind contains the delimiter", func() {
			BeforeEach(func() {
				indexName = "rode-v1alpha1-generic-resource"

				testFs[filepath.Join(config.MappingsPath, "generic-resource.json")] = mappingsFile(&VersionedMapping{
					Version:  fake.Word(),
					Mappings: map[string]interface{}{},
				})
			})

			It("should be parsed correctly", func() {
				Expect(actualIndexName).NotTo(BeNil())
				Expect(actualIndexName.Version).To(Equal("v1alpha1"))
				Expect(actualIndexName.DocumentKind).To(Equal("generic-resource"))
				Expect(actualIndexName.Inner).To(BeEmpty())
				Expect(actualOk).To(BeTrue())
			})
		})

		When("the inner name contains the delimiter", func() {
			BeforeEach(func() {
				indexName = "rode-v1alpha1-long-inner-name-generic-resource"

				testFs[filepath.Join(config.MappingsPath, "generic-resource.json")] = mappingsFile(&VersionedMapping{
					Version:  fake.Word(),
					Mappings: map[string]interface{}{},
				})
			})

			It("should be parsed correctly", func() {
				Expect(actualIndexName).NotTo(BeNil())
				Expect(actualIndexName.Version).To(Equal("v1alpha1"))
				Expect(actualIndexName.DocumentKind).To(Equal("generic-resource"))
				Expect(actualIndexName.Inner).To(Equal("long-inner-name"))
				Expect(actualOk).To(BeTrue())
			})
		})

		When("the document kind is unknown", func() {
			BeforeEach(func() {
				indexName = "rode-v1alpha1-foo"
			})

			It("should indicate that the document kind isn't in the registry", func() {
				Expect(actualIndexName).To(BeNil())
				Expect(actualOk).To(BeFalse())
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

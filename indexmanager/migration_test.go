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
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rode/es-index-manager/indexmanager"
	. "github.com/rode/es-index-manager/indexmanager/internal"
	"github.com/rode/es-index-manager/mocks"
)

var _ = Describe("Migrator", func() {
	var (
		ctx           = context.Background()
		config        *Config
		mockEsClient  *elasticsearch.Client
		mockTransport *mockEsTransport
		mockRegistry  *mocks.FakeMappingsRegistry
		mockRepo      *mocks.FakeIndexRepository

		documentKind        string
		expectedAlias       string
		expectedSourceIndex string
		expectedTargetIndex string
		expectedVersion     string
		expectedInnerName   string

		migrator Migrator

		expectedIndexPrefix string
	)

	BeforeEach(func() {
		expectedIndexPrefix = fake.Word()
		config = &Config{
			IndexPrefix: expectedIndexPrefix,
			Migration: &MigrationConfig{
				PollAttempts: 10,
				PollInterval: 10 * time.Second,
			},
		}

		mockTransport = &mockEsTransport{}
		mockEsClient = &elasticsearch.Client{Transport: mockTransport, API: esapi.New(mockTransport)}

		mockRegistry = &mocks.FakeMappingsRegistry{}
		mockRepo = &mocks.FakeIndexRepository{}
		fakeSleep := func(duration time.Duration) {
			Expect(duration).To(Equal(config.Migration.PollInterval))
		}

		expectedVersion = fake.Word()
		documentKind = fake.Word()
		expectedInnerName = fake.Word()

		expectedSourceIndex = createIndexOrAliasName(expectedIndexPrefix, fake.Word(), expectedInnerName, documentKind)
		expectedTargetIndex = createIndexOrAliasName(expectedIndexPrefix, expectedVersion, expectedInnerName, documentKind)
		expectedAlias = createIndexOrAliasName(expectedIndexPrefix, expectedInnerName, documentKind)

		migrator = NewMigrator(logger, mockEsClient, mockRegistry, mockRepo, fakeSleep, config)
	})

	Context("GetMigrations", func() {
		var (
			actualMigrations []*Migration
			actualError      error
		)

		BeforeEach(func() {
			mockTransport.preparedHttpResponses = []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: createESBody(map[string]interface{}{
						expectedSourceIndex: map[string]interface{}{
							"mappings": map[string]interface{}{
								"_meta": map[string]interface{}{
									"type": config.IndexPrefix,
								},
							},
						},
					}),
				},
			}

			mockRegistry.VersionReturns(expectedVersion)
			mockRegistry.AliasNameReturns(expectedAlias)
			mockRegistry.IndexNameReturns(expectedTargetIndex)
		})

		JustBeforeEach(func() {
			actualMigrations, actualError = migrator.GetMigrations(ctx)
		})

		When("there are pending migrations", func() {
			It("should fetch all indices from Elasticsearch", func() {
				Expect(actualError).To(BeNil())
				Expect(mockTransport.receivedHttpRequests[0].Method).To(Equal(http.MethodGet))
				Expect(mockTransport.receivedHttpRequests[0].URL.Path).To(Equal("/_all"))
			})

			It("should grab the current version for the document kind", func() {
				Expect(mockRegistry.VersionCallCount()).To(Equal(1))
				actualDocumentKind := mockRegistry.VersionArgsForCall(0)

				Expect(actualDocumentKind).To(Equal(documentKind))
			})

			It("should use the registry to get the target index and alias names", func() {
				Expect(mockRegistry.AliasNameCallCount()).To(Equal(1))
				actualDocumentKind, actualInnerName := mockRegistry.AliasNameArgsForCall(0)

				Expect(actualDocumentKind).To(Equal(documentKind))
				Expect(actualInnerName).To(Equal(expectedInnerName))

				Expect(mockRegistry.IndexNameCallCount()).To(Equal(1))
				actualDocumentKind, actualInnerName = mockRegistry.IndexNameArgsForCall(0)

				Expect(actualDocumentKind).To(Equal(actualDocumentKind))
				Expect(actualInnerName).To(Equal(actualInnerName))
			})

			It("should return the list of migrations", func() {
				Expect(actualMigrations).To(HaveLen(1))

				actualMigration := actualMigrations[0]
				Expect(actualMigration.DocumentKind).To(Equal(documentKind))
				Expect(actualMigration.SourceIndex).To(Equal(expectedSourceIndex))
				Expect(actualMigration.TargetIndex).To(Equal(expectedTargetIndex))
				Expect(actualMigration.Alias).To(Equal(expectedAlias))
			})
		})

		When("the index is up to date", func() {
			BeforeEach(func() {
				expectedSourceIndex = createIndexOrAliasName(expectedIndexPrefix, expectedVersion, expectedInnerName, documentKind)
				mockTransport.preparedHttpResponses[0].Body = createESBody(map[string]interface{}{
					expectedSourceIndex: map[string]interface{}{
						"mappings": map[string]interface{}{
							"_meta": map[string]interface{}{
								"type": config.IndexPrefix,
							},
						},
					},
				})
			})

			It("should not return any migrations", func() {
				Expect(actualMigrations).To(BeEmpty())
			})
		})

		When("no indices belong to the application", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[0].Body = createESBody(map[string]interface{}{
					fake.Word(): map[string]interface{}{
						"mappings": map[string]interface{}{},
					},
				})
			})

			It("shouldn't find any migrations", func() {
				Expect(actualMigrations).To(BeEmpty())
			})
		})

		When("an index shares the application's prefix but not the meta type", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[0].Body = createESBody(map[string]interface{}{
					fake.Word(): map[string]interface{}{
						"mappings": map[string]interface{}{
							"_meta": map[string]interface{}{
								"type": fake.Word(),
							},
						},
					},
				})
			})

			It("shouldn't find any migrations", func() {
				Expect(actualMigrations).To(BeEmpty())
			})
		})

		When("an index doesn't have _meta mapping", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[0].Body = createESBody(map[string]interface{}{
					expectedSourceIndex: map[string]interface{}{
						"mappings": map[string]interface{}{},
					},
				})
			})

			It("shouldn't find any migrations", func() {
				Expect(actualMigrations).To(BeEmpty())
			})
		})

		When("an error occurs fetching indices", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[0] = &http.Response{
					StatusCode: http.StatusInternalServerError,
				}
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualMigrations).To(BeNil())
			})
		})

		When("the indices get response is invalid", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[0].Body = createInvalidBody()
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualMigrations).To(BeNil())
			})
		})
	})

	Context("Migrate", func() {
		var (
			actualError error

			taskId string
		)

		BeforeEach(func() {
			taskId = fake.Word()
			mockTransport.preparedHttpResponses = []*http.Response{
				// get index settings
				{
					StatusCode: http.StatusOK,
					Body: createESBody(map[string]interface{}{
						expectedSourceIndex: EsSettingsResponse{
							Settings: &EsSettingsIndex{
								Index: &EsSettingsBlocks{
									Blocks: &EsSettingsWrite{
										Write: "false",
									},
								},
							},
						},
					}),
				},
				// add write block
				{
					StatusCode: http.StatusOK,
					Body: createESBody(&EsBlockResponse{
						Acknowledged:       true,
						ShardsAcknowledged: true,
					}),
				},
				// reindex
				{
					StatusCode: http.StatusOK,
					Body: createESBody(&EsTaskCreationResponse{
						Task: taskId,
					}),
				},
				// poll task
				{
					StatusCode: http.StatusOK,
					Body: createESBody(&EsTask{
						Completed: true,
					}),
				},
				// delete task document
				{
					StatusCode: http.StatusOK,
				},
				// update aliases
				{
					StatusCode: http.StatusOK,
				},
				// delete old index
				{
					StatusCode: http.StatusOK,
				},
			}
		})

		JustBeforeEach(func() {
			actualError = migrator.Migrate(ctx, &Migration{
				Alias:        expectedAlias,
				SourceIndex:  expectedSourceIndex,
				TargetIndex:  expectedTargetIndex,
				DocumentKind: documentKind,
			})
		})

		When("a successful migration occurs", func() {
			It("should not return an error", func() {
				Expect(actualError).NotTo(HaveOccurred())
			})

			It("should check if the source index already has a write block", func() {
				Expect(mockTransport.receivedHttpRequests[0].Method).To(Equal(http.MethodGet))
				Expect(mockTransport.receivedHttpRequests[0].URL.Path).To(Equal(fmt.Sprintf("/%s/_settings", expectedSourceIndex)))
			})

			It("should place a write block on the source index", func() {
				Expect(mockTransport.receivedHttpRequests[1].Method).To(Equal(http.MethodPut))
				Expect(mockTransport.receivedHttpRequests[1].URL.Path).To(Equal(fmt.Sprintf("/%s/_block/write", expectedSourceIndex)))
			})

			It("should create the target index", func() {
				Expect(mockRepo.CreateIndexCallCount()).To(Equal(1))

				_, actualIndex, actualAlias, actualDocumentKind := mockRepo.CreateIndexArgsForCall(0)
				Expect(actualIndex).To(Equal(expectedTargetIndex))
				Expect(actualAlias).To(Equal(expectedAlias))
				Expect(actualDocumentKind).To(Equal(documentKind))
			})

			It("should start a reindex on the source index to the target index", func() {
				expectedBody := &EsReindex{
					Conflicts: "proceed",
					Source: &EsReindexFields{
						Index: expectedSourceIndex,
					},
					Destination: &EsReindexFields{
						Index:  expectedTargetIndex,
						OpType: "create",
					},
				}
				actualBody := &EsReindex{}

				readRequestBody(mockTransport.receivedHttpRequests[2], actualBody)

				Expect(mockTransport.receivedHttpRequests[2].Method).To(Equal(http.MethodPost))
				Expect(mockTransport.receivedHttpRequests[2].URL.Path).To(Equal("/_reindex"))
				Expect(mockTransport.receivedHttpRequests[2].URL.Query().Get("wait_for_completion")).To(Equal("false"))
				Expect(actualBody).To(Equal(expectedBody))
			})

			It("should poll for the reindex task to complete", func() {
				Expect(mockTransport.receivedHttpRequests[3].Method).To(Equal(http.MethodGet))
				Expect(mockTransport.receivedHttpRequests[3].URL.Path).To(Equal("/_tasks/" + taskId))
			})

			It("should delete the task document once the reindex has finished", func() {
				Expect(mockTransport.receivedHttpRequests[4].Method).To(Equal(http.MethodDelete))
				Expect(mockTransport.receivedHttpRequests[4].URL.Path).To(Equal(fmt.Sprintf("/%s/_doc/%s", ".tasks", taskId)))
			})

			It("should point the alias to the target index", func() {
				expectedBody := &EsIndexAliasRequest{
					Actions: []EsActions{
						{
							Remove: &EsIndexAlias{
								Index: expectedSourceIndex,
								Alias: expectedAlias,
							},
						},
						{
							Add: &EsIndexAlias{
								Index: expectedTargetIndex,
								Alias: expectedAlias,
							},
						},
					},
				}
				actualBody := &EsIndexAliasRequest{}
				readRequestBody(mockTransport.receivedHttpRequests[5], actualBody)

				Expect(mockTransport.receivedHttpRequests[5].Method).To(Equal(http.MethodPost))
				Expect(mockTransport.receivedHttpRequests[5].URL.Path).To(Equal("/_aliases"))
				Expect(actualBody).To(Equal(expectedBody))
			})

			It("should delete the source index", func() {
				Expect(mockTransport.receivedHttpRequests[6].Method).To(Equal(http.MethodDelete))
				Expect(mockTransport.receivedHttpRequests[6].URL.Path).To(Equal("/" + expectedSourceIndex))
			})
		})

		Context("placing the write block fails", func() {
			When("an error occurs checking if the source index has a write block", func() {
				BeforeEach(func() {
					mockTransport.preparedHttpResponses[0].StatusCode = http.StatusInternalServerError
				})

				It("should return an error and not make any additional requests", func() {
					Expect(actualError).NotTo(BeNil())
					Expect(actualError.Error()).To(ContainSubstring("error checking if write block is enabled on index"))

					Expect(mockTransport.receivedHttpRequests).To(HaveLen(1))
				})
			})

			When("the settings response body is invalid", func() {
				BeforeEach(func() {
					mockTransport.preparedHttpResponses[0].Body = createInvalidBody()
				})

				It("should return an error and not make any additional requests", func() {
					Expect(actualError).NotTo(BeNil())
					Expect(actualError.Error()).To(ContainSubstring("error decoding settings response"))

					Expect(mockTransport.receivedHttpRequests).To(HaveLen(1))
				})
			})

			When("an error occurs placing the write block on the source index", func() {
				BeforeEach(func() {
					mockTransport.preparedHttpResponses[1].StatusCode = http.StatusInternalServerError
				})

				It("should return an error and not make any additional requests", func() {
					Expect(actualError).NotTo(BeNil())
					Expect(actualError.Error()).To(ContainSubstring("error placing write block on index"))

					Expect(mockTransport.receivedHttpRequests).To(HaveLen(2))
				})
			})

			When("the write block response body is invalid", func() {
				BeforeEach(func() {
					mockTransport.preparedHttpResponses[1].Body = createInvalidBody()
				})

				It("should return an error and not make any additional requests", func() {
					Expect(actualError).NotTo(BeNil())
					Expect(actualError.Error()).To(ContainSubstring("error decoding write block response"))

					Expect(mockTransport.receivedHttpRequests).To(HaveLen(2))
				})
			})

			When("the write block isn't acknowledged", func() {
				BeforeEach(func() {
					mockTransport.preparedHttpResponses[1].Body = createESBody(&EsBlockResponse{
						Acknowledged: false,
					})
				})

				It("should return an error and not make any additional requests", func() {
					Expect(actualError).NotTo(BeNil())
					Expect(actualError.Error()).To(ContainSubstring("unable to block writes"))

					Expect(mockTransport.receivedHttpRequests).To(HaveLen(2))
				})
			})

			When("the write block isn't acknowledged by all shards", func() {
				BeforeEach(func() {
					mockTransport.preparedHttpResponses[1].Body = createESBody(&EsBlockResponse{
						Acknowledged:       true,
						ShardsAcknowledged: false,
					})
				})

				It("should return an error and not make any additional requests", func() {
					Expect(actualError).NotTo(BeNil())
					Expect(actualError.Error()).To(ContainSubstring("unable to block writes"))

					Expect(mockTransport.receivedHttpRequests).To(HaveLen(2))
				})
			})
		})

		When("the source index already has a write block", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[0].Body = createESBody(map[string]interface{}{
					expectedSourceIndex: EsSettingsResponse{
						Settings: &EsSettingsIndex{
							Index: &EsSettingsBlocks{
								Blocks: &EsSettingsWrite{
									Write: "true",
								},
							},
						},
					},
				})

				mockTransport.preparedHttpResponses = append(
					mockTransport.preparedHttpResponses[:1],
					mockTransport.preparedHttpResponses[2:]...)
			})

			It("should not try to place a block", func() {
				Expect(mockTransport.receivedHttpRequests).To(HaveLen(6))
			})
		})

		When("creating the target index fails", func() {
			BeforeEach(func() {
				mockRepo.CreateIndexReturns(errors.New(fake.Word()))
			})

			It("should return an error and not make any additional requests", func() {
				Expect(actualError).NotTo(BeNil())
				Expect(actualError.Error()).To(ContainSubstring("error creating target index"))

				Expect(mockTransport.receivedHttpRequests).To(HaveLen(2))
			})
		})

		When("the reindex request is unsuccessful", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[2].StatusCode = http.StatusInternalServerError
			})

			It("should return an error and not make any additional requests", func() {
				Expect(actualError).NotTo(BeNil())
				Expect(actualError.Error()).To(ContainSubstring("error initiating reindex"))

				Expect(mockTransport.receivedHttpRequests).To(HaveLen(3))
			})
		})

		When("the reindex response is invalid", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[2].Body = createInvalidBody()
			})

			It("should return an error and not make any additional requests", func() {
				Expect(actualError).NotTo(BeNil())
				Expect(actualError.Error()).To(ContainSubstring("error decoding reindex response"))

				Expect(mockTransport.receivedHttpRequests).To(HaveLen(3))
			})
		})

		When("the reindexing doesn't complete", func() {
			BeforeEach(func() {
				responses := make([]*http.Response, 13)

				for i := 0; i < len(responses); i++ {
					if i >= 3 {
						responses[i] = &http.Response{
							StatusCode: http.StatusOK,
							Body: createESBody(&EsTask{
								Completed: false,
							}),
						}
					} else {
						responses[i] = mockTransport.preparedHttpResponses[i]
					}
				}

				mockTransport.preparedHttpResponses = responses
			})

			It("should return an error and not make any additional requests", func() {
				Expect(actualError).NotTo(BeNil())
				Expect(actualError.Error()).To(ContainSubstring("reindex did not complete"))

				Expect(mockTransport.receivedHttpRequests).To(HaveLen(13))
			})
		})

		When("fetching the task fails once", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[3].StatusCode = http.StatusInternalServerError
				taskCompletedResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body: createESBody(&EsTask{
						Completed: true,
					}),
				}
				mockTransport.preparedHttpResponses = insertResponseAt(mockTransport.preparedHttpResponses, taskCompletedResponse, 4)
			})

			It("should continue and not return an error", func() {
				Expect(actualError).To(BeNil())
			})
		})

		When("the task response is invalid", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[3].Body = createInvalidBody()

				taskCompletedResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body: createESBody(&EsTask{
						Completed: true,
					}),
				}
				mockTransport.preparedHttpResponses = insertResponseAt(mockTransport.preparedHttpResponses, taskCompletedResponse, 4)
			})

			It("should continue and not return an error", func() {
				Expect(actualError).To(BeNil())
			})
		})

		When("an error occurs deleting the task document", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[4].StatusCode = http.StatusInternalServerError
			})

			It("should not return an error", func() {
				Expect(actualError).To(BeNil())
			})
		})

		When("an error occurs updating the alias", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[5].StatusCode = http.StatusInternalServerError
			})

			It("should return an error", func() {
				Expect(actualError).NotTo(BeNil())
				Expect(actualError.Error()).To(ContainSubstring("error occurred while swapping the alias"))
			})
		})

		When("an error occurs deleting the source index", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses[6].StatusCode = http.StatusInternalServerError
			})

			It("should return an error", func() {
				Expect(actualError).NotTo(BeNil())
				Expect(actualError.Error()).To(ContainSubstring("failed to remove the source index"))
			})
		})
	})
})

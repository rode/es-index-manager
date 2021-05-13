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
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/rode/es-index-manager/indexmanager"
	"github.com/rode/es-index-manager/mocks"
)

var _ = Describe("IndexRepository", func() {
	var (
		registry      *mocks.FakeMappingsRegistry
		mockTransport *mockEsTransport
		mockEsClient  *elasticsearch.Client
		repository    IndexRepository

		ctx       = context.Background()
		indexName string
	)

	BeforeEach(func() {
		indexName = fake.Word()

		registry = &mocks.FakeMappingsRegistry{}
		mockTransport = &mockEsTransport{}
		mockEsClient = &elasticsearch.Client{Transport: mockTransport, API: esapi.New(mockTransport)}

		repository = NewIndexRepository(logger, mockEsClient, registry)
	})

	Context("CreateIndex", func() {
		var (
			documentKind string
			aliasName    string

			actualError     error
			expectedMapping *VersionedMapping
		)

		BeforeEach(func() {
			documentKind = fake.Word()
			aliasName = fake.Word()

			expectedMapping = createRandomMapping()
			registry.MappingReturns(expectedMapping)
		})

		JustBeforeEach(func() {
			actualError = repository.CreateIndex(ctx, indexName, aliasName, documentKind)
		})

		When("the index does not exist", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses = []*http.Response{
					{
						StatusCode: http.StatusNotFound,
					},
					{
						StatusCode: http.StatusOK,
					},
				}
			})

			It("should not return an error", func() {
				Expect(actualError).To(BeNil())
			})

			It("should create the index", func() {
				Expect(mockTransport.receivedHttpRequests).To(HaveLen(2))
				Expect(mockTransport.receivedHttpRequests[1].Method).To(Equal(http.MethodPut))
				Expect(mockTransport.receivedHttpRequests[1].URL.Path).To(Equal("/" + indexName))
			})

			It("should pass the correct mappings", func() {
				actualPayload := map[string]interface{}{}

				readRequestBody(mockTransport.receivedHttpRequests[1], &actualPayload)

				Expect(actualPayload).To(MatchAllKeys(Keys{
					"mappings": Equal(expectedMapping.Mappings),
					"aliases": MatchAllKeys(Keys{
						aliasName: BeEmpty(),
					}),
				}))
			})
		})

		When("an unexpected status code while checking if the index exists", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses = []*http.Response{
					{
						StatusCode: http.StatusInternalServerError,
					},
				}
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualError.Error()).To(ContainSubstring("unexpected status code"))
			})
		})

		When("an error occurs checking if the index exists", func() {
			var (
				expectedError error
			)

			BeforeEach(func() {
				expectedError = errors.New(fake.Word())
				mockTransport.actions = []transportAction{
					func(req *http.Request) (*http.Response, error) {
						return nil, expectedError
					},
				}
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualError.Error()).To(ContainSubstring("error checking if index " + indexName + " exists"))
			})
		})

		When("an unexpected status code while creating the index", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses = []*http.Response{
					{
						StatusCode: http.StatusNotFound,
					},
					{
						StatusCode: http.StatusInternalServerError,
					},
				}
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualError.Error()).To(ContainSubstring("unexpected status code after creating index"))
			})
		})

		When("an error occurs creating the index", func() {
			var (
				expectedError error
			)

			BeforeEach(func() {
				expectedError = errors.New(fake.Word())
				mockTransport.actions = []transportAction{
					func(req *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusNotFound}, nil
					},
					func(req *http.Request) (*http.Response, error) {
						return nil, expectedError
					},
				}
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualError.Error()).To(ContainSubstring("error creating index"))
			})
		})

		When("the index already exists", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses = []*http.Response{
					{
						StatusCode: http.StatusOK,
					},
				}
			})

			It("should not try to create the index", func() {
				Expect(actualError).NotTo(HaveOccurred())
				Expect(mockTransport.receivedHttpRequests).To(HaveLen(1))
			})
		})

		When("the index is created by another caller", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses = []*http.Response{
					{
						StatusCode: http.StatusNotFound,
					},
					{
						StatusCode: http.StatusBadRequest,
						Body:       createEsErrorResponse("resource_already_exists_exception"),
					},
				}
			})

			It("should not return an error", func() {
				Expect(actualError).To(BeNil())
			})
		})

		When("the document kind isn't in the registry", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses = []*http.Response{
					{
						StatusCode: http.StatusNotFound,
					},
				}
				registry.MappingReturns(nil)
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualError.Error()).To(ContainSubstring("unable to find a mapping"))
			})
		})
	})

	Context("DeleteIndex", func() {
		var (
			actualError error
		)

		JustBeforeEach(func() {
			actualError = repository.DeleteIndex(ctx, indexName)
		})

		When("the index exists", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses = []*http.Response{
					{
						StatusCode: http.StatusOK,
					},
				}
			})

			It("should send a delete request", func() {
				Expect(mockTransport.receivedHttpRequests).To(HaveLen(1))
				Expect(mockTransport.receivedHttpRequests[0].URL.Path).To(Equal("/" + indexName))
				Expect(mockTransport.receivedHttpRequests[0].Method).To(Equal(http.MethodDelete))
			})

			It("should not return an error", func() {
				Expect(actualError).NotTo(HaveOccurred())
			})
		})

		When("an error occurs deleting the index", func() {
			BeforeEach(func() {
				mockTransport.actions = []transportAction{
					func(req *http.Request) (*http.Response, error) {
						return nil, errors.New(fake.Word())
					},
				}
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualError.Error()).To(ContainSubstring("error deleting index"))
			})
		})

		When("an unexpected response is returned", func() {
			BeforeEach(func() {
				mockTransport.preparedHttpResponses = []*http.Response{
					{
						StatusCode: http.StatusInternalServerError,
					},
				}
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualError.Error()).To(ContainSubstring("unexpected response"))
			})
		})
	})
})

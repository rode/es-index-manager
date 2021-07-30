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

package indexmanager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	. "github.com/rode/es-index-manager/indexmanager/internal"
	"go.uber.org/zap"
)

//counterfeiter:generate -o ../mocks . IndexRepository
type IndexRepository interface {
	// CreateIndex makes a new index using the mappings supplied for the document kind.
	// If there's an alias specified in metadata, it's added to the index.
	CreateIndex(ctx context.Context, indexName, aliasName, documentKind string) error
	// DeleteIndex deletes the index, which also removes any associated aliases.
	DeleteIndex(ctx context.Context, indexName string) error
}

type indexRepository struct {
	client   *elasticsearch.Client
	logger   *zap.Logger
	registry MappingsRegistry
}

func NewIndexRepository(logger *zap.Logger, client *elasticsearch.Client, registry MappingsRegistry) IndexRepository {
	return &indexRepository{
		client,
		logger,
		registry,
	}
}

func (ir *indexRepository) CreateIndex(ctx context.Context, indexName, aliasName, documentKind string) error {
	log := ir.logger.Named("CreateIndex").With(zap.String("index", indexName))

	res, err := ir.client.Indices.Exists([]string{indexName}, ir.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("error checking if index %s exists: %s", indexName, err)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
		log.Error("error checking if index exists", zap.String("response", res.String()))

		return fmt.Errorf("unexpected status code (%d) when checking if index exists", res.StatusCode)
	}

	// if the response was successful, then the index already exists and we can skip creation
	if !res.IsError() {
		return nil
	}

	mapping := ir.registry.Mapping(documentKind)
	if mapping == nil {
		return fmt.Errorf("unable to find a mapping for document kind %s", documentKind)
	}

	createIndexReq := map[string]interface{}{
		"mappings": mapping.Mappings,
	}
	if mapping.Settings != nil {
		createIndexReq["settings"] = mapping.Settings
	}

	if aliasName != "" {
		createIndexReq["aliases"] = map[string]interface{}{
			aliasName: map[string]interface{}{},
		}
	}

	payload, _ := encodeRequest(&createIndexReq)
	res, err = ir.client.Indices.Create(indexName, ir.client.Indices.Create.WithContext(ctx), ir.client.Indices.Create.WithBody(payload))
	if err != nil {
		return fmt.Errorf("error creating index %s: %s", indexName, err)
	}

	if res.IsError() {
		if res.StatusCode == http.StatusBadRequest {
			errResponse := EsErrorResponse{}
			if err := decodeResponse(res.Body, &errResponse); err != nil {
				return fmt.Errorf("error decoding Elasticsearch error response: %s", err)
			}

			// there's a chance for another instance of the application to try to create the same index (e.g., during migrations)
			// so treat that differently than an error
			if errResponse.Error.Type == ElasticsearchResourceAlreadyExists {
				log.Info("index already exists")
				return nil
			}
		}

		return fmt.Errorf("unexpected status code after creating index: %d", res.StatusCode)
	}

	log.Info("index created")

	return nil
}

func (ir *indexRepository) DeleteIndex(ctx context.Context, indexName string) error {
	log := ir.logger.Named("DeleteIndex").With(zap.String("index", indexName))

	res, err := ir.client.Indices.Delete(
		[]string{indexName},
		ir.client.Indices.Delete.WithContext(ctx),
	)

	if err != nil {
		return fmt.Errorf("error deleting index %s: %s", indexName, err)
	}

	if res.IsError() {
		return fmt.Errorf("unexpected response from elasticsearch: %s", res.String())
	}

	log.Debug("index deleted")

	return nil
}

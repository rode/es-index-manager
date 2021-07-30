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
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	. "github.com/rode/es-index-manager/indexmanager/internal"
	"go.uber.org/zap"
)

type migrator struct {
	config   *Config
	client   *elasticsearch.Client
	logger   *zap.Logger
	registry MappingsRegistry
	repo     IndexRepository
	sleep    func(time.Duration)
}

//counterfeiter:generate -o ../mocks . Migrator
type Migrator interface {
	GetMigrations(ctx context.Context) ([]*Migration, error)
	Migrate(ctx context.Context, migration *Migration) error
}

func NewMigrator(
	logger *zap.Logger,
	client *elasticsearch.Client,
	registry MappingsRegistry,
	repo IndexRepository,
	sleep func(time.Duration),
	config *Config,
) Migrator {
	return &migrator{
		config,
		client,
		logger,
		registry,
		repo,
		sleep,
	}
}

func (m *migrator) GetMigrations(ctx context.Context) ([]*Migration, error) {
	log := m.logger.Named("GetMigrations")
	res, err := m.client.Indices.Get([]string{ElasticsearchAllIndices}, m.client.Indices.Get.WithContext(ctx))
	if err := getErrorFromESResponse(res, err); err != nil {
		return nil, err
	}

	allIndices := map[string]EsIndex{}

	if err := decodeResponse(res.Body, &allIndices); err != nil {
		return nil, err
	}

	var migrations []*Migration
	for indexName, indexValue := range allIndices {
		meta := indexValue.Mappings.Meta
		if !(strings.HasPrefix(indexName, m.config.IndexPrefix) && meta != nil && meta.Type == m.config.IndexPrefix) {
			continue
		}

		indexParts, ok := m.registry.ParseIndexName(indexName)
		if !ok {
			log.Warn("Discovered index matching criteria, but wasn't able to determine document kind.", zap.String("index", indexName))
			continue
		}

		currentVersion := m.registry.Version(indexParts.DocumentKind)

		if indexParts.Version == currentVersion {
			continue
		}

		alias := m.registry.AliasName(indexParts.DocumentKind, indexParts.Inner)
		targetIndex := m.registry.IndexName(indexParts.DocumentKind, indexParts.Inner)
		migration := &Migration{
			SourceIndex:  indexName,
			TargetIndex:  targetIndex,
			DocumentKind: indexParts.DocumentKind,
			Alias:        alias,
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

func (m *migrator) Migrate(ctx context.Context, migration *Migration) error {
	log := m.logger.Named("Migrate").
		With(zap.String("source", migration.SourceIndex)).
		With(zap.String("target", migration.TargetIndex))

	log.Info("Starting migration")

	if err := m.blockWritesOnIndex(ctx, log, migration.SourceIndex); err != nil {
		return err
	}

	err := m.repo.CreateIndex(ctx, migration.TargetIndex, migration.Alias, migration.DocumentKind)
	if err != nil {
		return fmt.Errorf("error creating target index: %s", err)
	}

	if err := m.reindex(ctx, log, migration.SourceIndex, migration.TargetIndex); err != nil {
		return err
	}

	if err := m.swapAlias(ctx, log, migration.Alias, migration.SourceIndex, migration.TargetIndex); err != nil {
		return err
	}

	log.Info("Deleting source index")
	res, err := m.client.Indices.Delete(
		[]string{migration.SourceIndex},
		m.client.Indices.Delete.WithContext(ctx),
	)

	if err != nil {
		return fmt.Errorf("failed to remove source index: %s", err)
	}

	if res.IsError() && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to remove the source index, status: %d", res.StatusCode)
	}

	log.Info("Migration complete")
	return nil
}

func (m *migrator) blockWritesOnIndex(ctx context.Context, log *zap.Logger, indexName string) error {
	res, err := m.client.Indices.GetSettings(m.client.Indices.GetSettings.WithContext(ctx), m.client.Indices.GetSettings.WithIndex(indexName))
	if err := getErrorFromESResponse(res, err); err != nil {
		return fmt.Errorf("error checking if write block is enabled on index: %s", err)
	}

	settingsResponse := map[string]EsSettingsResponse{}
	if err := decodeResponse(res.Body, &settingsResponse); err != nil {
		return fmt.Errorf("error decoding settings response: %s", err)
	}

	blocks := settingsResponse[indexName].Settings.Index.Blocks

	// index already has a write block in place
	if blocks != nil && blocks.Write == "true" {
		return nil
	}

	log.Info("Placing write block on index")
	res, err = m.client.Indices.AddBlock([]string{indexName}, "write", m.client.Indices.AddBlock.WithContext(ctx))
	if err := getErrorFromESResponse(res, err); err != nil {
		return fmt.Errorf("error placing write block on index: %s", err)
	}

	blockResponse := &EsBlockResponse{}
	if err := decodeResponse(res.Body, blockResponse); err != nil {
		return fmt.Errorf("error decoding write block response: %s", err)
	}

	if !(blockResponse.Acknowledged && blockResponse.ShardsAcknowledged) {
		log.Error("Write block unsuccessful", zap.Any("response", blockResponse))
		return fmt.Errorf("unable to block writes for index: %s", indexName)
	}

	return nil
}

func (m *migrator) reindex(ctx context.Context, log *zap.Logger, sourceIndex, targetIndex string) error {
	reindexReq := &EsReindex{
		Conflicts:   "proceed",
		Source:      &EsReindexFields{Index: sourceIndex},
		Destination: &EsReindexFields{Index: targetIndex, OpType: "create"},
	}
	reindexBody, _ := encodeRequest(reindexReq)
	log.Info("Starting reindex")
	res, err := m.client.Reindex(
		reindexBody,
		m.client.Reindex.WithContext(ctx),
		m.client.Reindex.WithWaitForCompletion(false))
	if err := getErrorFromESResponse(res, err); err != nil {
		return fmt.Errorf("error initiating reindex: %s", err)
	}
	taskCreationResponse := &EsTaskCreationResponse{}

	if err := decodeResponse(res.Body, taskCreationResponse); err != nil {
		return fmt.Errorf("error decoding reindex response: %s", err)
	}
	log.Info("Reindex started", zap.String("taskId", taskCreationResponse.Task))

	reindexCompleted := false
	for i := 0; i < m.config.Migration.PollAttempts; i++ {
		log.Info("Polling task API", zap.String("taskId", taskCreationResponse.Task))
		res, err = m.client.Tasks.Get(taskCreationResponse.Task, m.client.Tasks.Get.WithContext(ctx))
		if err := getErrorFromESResponse(res, err); err != nil {
			log.Warn("error getting task status", zap.Error(err))
			continue
		}

		task := &EsTask{}
		if err := decodeResponse(res.Body, task); err != nil {
			log.Warn("error decoding task response", zap.Error(err))
			continue
		}

		if task.Completed {
			reindexCompleted = true
			log.Info("Reindex completed")

			break
		}

		log.Info("Task incomplete, waiting before polling again", zap.String("taskId", taskCreationResponse.Task))
		m.sleep(m.config.Migration.PollInterval)
	}

	if !reindexCompleted {
		return fmt.Errorf("reindex did not complete after %d polls", m.config.Migration.PollAttempts)
	}

	res, err = m.client.Delete(ElasticsearchTaskIndex, taskCreationResponse.Task, m.client.Delete.WithContext(ctx))
	if err := getErrorFromESResponse(res, err); err != nil {
		log.Warn("Error deleting task document", zap.Error(err), zap.String("taskId", taskCreationResponse.Task))
	}

	return nil
}

func (m *migrator) swapAlias(ctx context.Context, log *zap.Logger, alias, sourceIndex, targetIndex string) error {
	log = log.With(zap.String("alias", alias))

	aliasReq := &EsIndexAliasRequest{
		Actions: []EsActions{
			{
				Remove: &EsIndexAlias{
					Index: sourceIndex,
					Alias: alias,
				},
			},
			{
				Add: &EsIndexAlias{
					Index: targetIndex,
					Alias: alias,
				},
			},
		},
	}

	aliasReqBody, _ := encodeRequest(aliasReq)
	log.Info("Swapping alias over to new index")
	res, err := m.client.Indices.UpdateAliases(
		aliasReqBody,
		m.client.Indices.UpdateAliases.WithContext(ctx),
	)

	if err := getErrorFromESResponse(res, err); err != nil {
		return fmt.Errorf("error occurred while swapping the alias: %s", err)
	}

	return nil
}

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
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/zap"
)

//go:generate counterfeiter -generate

//counterfeiter:generate -o ../mocks . IndexManager
type IndexManager interface {
	MappingsRegistry
	IndexRepository
	MigrationOrchestrator
	Initialize(context.Context) error
}

type indexManager struct {
	MappingsRegistry
	IndexRepository
	MigrationOrchestrator
}

func NewIndexManager(logger *zap.Logger, client *elasticsearch.Client, config *Config) IndexManager {
	if config.Migration == nil {
		config.Migration = &MigrationConfig{
			PollAttempts: 10,
			PollInterval: time.Second * 10,
		}
	}

	registry := NewMappingsRegistry(config, os.DirFS("."))
	repo := NewIndexRepository(logger, client, registry)
	orchestrator := NewMigrationOrchestrator(logger, NewMigrator(logger, client, registry, repo, config))
	return &indexManager{
		registry,
		repo,
		orchestrator,
	}
}

func (im *indexManager) Initialize(ctx context.Context) error {
	if err := im.LoadMappings(); err != nil {
		return fmt.Errorf("error occurred loading index mappings: %s", err)
	}

	if err := im.RunMigrations(ctx); err != nil {
		return fmt.Errorf("error running migrations: %s", err)
	}

	return nil
}

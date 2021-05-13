package indexmanager

import (
	"context"
	"fmt"
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

	registry := NewMappingsRegistry(config)
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

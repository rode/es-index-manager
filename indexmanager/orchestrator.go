package indexmanager

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

//counterfeiter:generate -o ../mocks . MigrationOrchestrator
type MigrationOrchestrator interface {
	RunMigrations(ctx context.Context) error
}

type migrationOrchestrator struct {
	logger   *zap.Logger
	migrator Migrator
}

func NewMigrationOrchestrator(logger *zap.Logger, migrator Migrator) MigrationOrchestrator {
	return &migrationOrchestrator{
		logger:   logger,
		migrator: migrator,
	}
}

func (m *migrationOrchestrator) RunMigrations(ctx context.Context) error {
	log := m.logger.Named("RunMigrations")
	migrationsToRun, err := m.migrator.GetMigrations(ctx)
	if err != nil {
		return err
	}

	if len(migrationsToRun) == 0 {
		log.Info("No migrations to run")
		return nil
	}

	log.Info(fmt.Sprintf("Discovered %d migrations to run", len(migrationsToRun)))

	for _, migration := range migrationsToRun {
		if err := m.migrator.Migrate(ctx, migration); err != nil {
			return err
		}
	}

	return nil
}

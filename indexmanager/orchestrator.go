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
	migrations, err := m.migrator.GetMigrations(ctx)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		log.Info("No migrations to run")
		return nil
	}

	log.Info(fmt.Sprintf("Discovered %d migrations to run", len(migrations)))

	for _, migration := range migrations {
		if err := m.migrator.Migrate(ctx, migration); err != nil {
			return err
		}
	}

	return nil
}

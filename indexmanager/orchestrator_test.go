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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rode/es-index-manager/indexmanager"
	"github.com/rode/es-index-manager/mocks"
)

var _ = Describe("MigrationOrchestrator", func() {

	var (
		ctx          = context.Background()
		mockMigrator *mocks.FakeMigrator
		orchestrator MigrationOrchestrator
	)

	BeforeEach(func() {
		mockMigrator = &mocks.FakeMigrator{}
		orchestrator = NewMigrationOrchestrator(logger, mockMigrator)
	})

	Context("RunMigrations", func() {
		var (
			getMigrationsError error
			migrations         []*Migration

			actualError error
		)

		BeforeEach(func() {
			getMigrationsError = nil
			migrations = []*Migration{}

			for i := 0; i < fake.Number(2, 5); i++ {
				migrations = append(migrations, &Migration{
					SourceIndex:  fake.Word(),
					TargetIndex:  fake.Word(),
					DocumentKind: fake.Word(),
				})
			}
		})

		JustBeforeEach(func() {
			mockMigrator.GetMigrationsReturns(migrations, getMigrationsError)

			actualError = orchestrator.RunMigrations(ctx)
		})

		When("there are no migrations", func() {
			BeforeEach(func() {
				migrations = []*Migration{}
			})

			It("should try to fetch any pending migrations", func() {
				Expect(mockMigrator.GetMigrationsCallCount()).To(Equal(1))
			})

			It("should not call migrate", func() {
				Expect(mockMigrator.MigrateCallCount()).To(Equal(0))
			})

			It("should not return an error", func() {
				Expect(actualError).To(BeNil())
			})
		})

		When("there are multiple migrations", func() {
			It("should call the migrator for each", func() {
				Expect(mockMigrator.MigrateCallCount()).To(Equal(len(migrations)))

				for i := 0; i < len(migrations); i++ {
					_, actualMigration := mockMigrator.MigrateArgsForCall(0)

					Expect(actualMigration).To(BeElementOf(migrations))
				}
			})

			It("should not return an error", func() {
				Expect(actualError).To(BeNil())
			})
		})

		When("an error occurs discovering migrations to run", func() {
			BeforeEach(func() {
				getMigrationsError = errors.New(fake.Word())
			})

			It("Should return an error", func() {
				Expect(actualError).NotTo(BeNil())
			})
		})

		When("a migration fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New(fake.Word())

				mockMigrator.MigrateReturns(expectedError)
			})

			It("should return the error", func() {
				Expect(actualError).To(Equal(expectedError))
			})
		})
	})
})

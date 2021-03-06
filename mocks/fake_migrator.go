// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"context"
	"sync"

	"github.com/rode/es-index-manager/indexmanager"
)

type FakeMigrator struct {
	GetMigrationsStub        func(context.Context) ([]*indexmanager.Migration, error)
	getMigrationsMutex       sync.RWMutex
	getMigrationsArgsForCall []struct {
		arg1 context.Context
	}
	getMigrationsReturns struct {
		result1 []*indexmanager.Migration
		result2 error
	}
	getMigrationsReturnsOnCall map[int]struct {
		result1 []*indexmanager.Migration
		result2 error
	}
	MigrateStub        func(context.Context, *indexmanager.Migration) error
	migrateMutex       sync.RWMutex
	migrateArgsForCall []struct {
		arg1 context.Context
		arg2 *indexmanager.Migration
	}
	migrateReturns struct {
		result1 error
	}
	migrateReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeMigrator) GetMigrations(arg1 context.Context) ([]*indexmanager.Migration, error) {
	fake.getMigrationsMutex.Lock()
	ret, specificReturn := fake.getMigrationsReturnsOnCall[len(fake.getMigrationsArgsForCall)]
	fake.getMigrationsArgsForCall = append(fake.getMigrationsArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.GetMigrationsStub
	fakeReturns := fake.getMigrationsReturns
	fake.recordInvocation("GetMigrations", []interface{}{arg1})
	fake.getMigrationsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeMigrator) GetMigrationsCallCount() int {
	fake.getMigrationsMutex.RLock()
	defer fake.getMigrationsMutex.RUnlock()
	return len(fake.getMigrationsArgsForCall)
}

func (fake *FakeMigrator) GetMigrationsCalls(stub func(context.Context) ([]*indexmanager.Migration, error)) {
	fake.getMigrationsMutex.Lock()
	defer fake.getMigrationsMutex.Unlock()
	fake.GetMigrationsStub = stub
}

func (fake *FakeMigrator) GetMigrationsArgsForCall(i int) context.Context {
	fake.getMigrationsMutex.RLock()
	defer fake.getMigrationsMutex.RUnlock()
	argsForCall := fake.getMigrationsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeMigrator) GetMigrationsReturns(result1 []*indexmanager.Migration, result2 error) {
	fake.getMigrationsMutex.Lock()
	defer fake.getMigrationsMutex.Unlock()
	fake.GetMigrationsStub = nil
	fake.getMigrationsReturns = struct {
		result1 []*indexmanager.Migration
		result2 error
	}{result1, result2}
}

func (fake *FakeMigrator) GetMigrationsReturnsOnCall(i int, result1 []*indexmanager.Migration, result2 error) {
	fake.getMigrationsMutex.Lock()
	defer fake.getMigrationsMutex.Unlock()
	fake.GetMigrationsStub = nil
	if fake.getMigrationsReturnsOnCall == nil {
		fake.getMigrationsReturnsOnCall = make(map[int]struct {
			result1 []*indexmanager.Migration
			result2 error
		})
	}
	fake.getMigrationsReturnsOnCall[i] = struct {
		result1 []*indexmanager.Migration
		result2 error
	}{result1, result2}
}

func (fake *FakeMigrator) Migrate(arg1 context.Context, arg2 *indexmanager.Migration) error {
	fake.migrateMutex.Lock()
	ret, specificReturn := fake.migrateReturnsOnCall[len(fake.migrateArgsForCall)]
	fake.migrateArgsForCall = append(fake.migrateArgsForCall, struct {
		arg1 context.Context
		arg2 *indexmanager.Migration
	}{arg1, arg2})
	stub := fake.MigrateStub
	fakeReturns := fake.migrateReturns
	fake.recordInvocation("Migrate", []interface{}{arg1, arg2})
	fake.migrateMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeMigrator) MigrateCallCount() int {
	fake.migrateMutex.RLock()
	defer fake.migrateMutex.RUnlock()
	return len(fake.migrateArgsForCall)
}

func (fake *FakeMigrator) MigrateCalls(stub func(context.Context, *indexmanager.Migration) error) {
	fake.migrateMutex.Lock()
	defer fake.migrateMutex.Unlock()
	fake.MigrateStub = stub
}

func (fake *FakeMigrator) MigrateArgsForCall(i int) (context.Context, *indexmanager.Migration) {
	fake.migrateMutex.RLock()
	defer fake.migrateMutex.RUnlock()
	argsForCall := fake.migrateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeMigrator) MigrateReturns(result1 error) {
	fake.migrateMutex.Lock()
	defer fake.migrateMutex.Unlock()
	fake.MigrateStub = nil
	fake.migrateReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeMigrator) MigrateReturnsOnCall(i int, result1 error) {
	fake.migrateMutex.Lock()
	defer fake.migrateMutex.Unlock()
	fake.MigrateStub = nil
	if fake.migrateReturnsOnCall == nil {
		fake.migrateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.migrateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeMigrator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getMigrationsMutex.RLock()
	defer fake.getMigrationsMutex.RUnlock()
	fake.migrateMutex.RLock()
	defer fake.migrateMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeMigrator) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ indexmanager.Migrator = new(FakeMigrator)

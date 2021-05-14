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
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

//counterfeiter:generate -o ../mocks . MappingsRegistry
type MappingsRegistry interface {
	// LoadMappings reads the index mapping and version from JSON file in the directory specified in Config.MappingsPath.
	LoadMappings() error
	// IndexName returns the full index name, using the prefix, version, inner name, and document kind.
	IndexName(documentKind, inner string) string
	// AliasName returns the full alias name, using the prefix, inner name, and document kind.
	AliasName(documentKind, inner string) string
	// Version returns the current schema version for the given document kind.
	Version(documentKind string) string
	// Mapping returns the current versioned mapping for the document kind.
	Mapping(documentKind string) *VersionedMapping
	// ParseIndexName determines the version, document kind, and inner name of an index.
	ParseIndexName(indexName string) *IndexName
}

type mappingsRegistry struct {
	config     *Config
	filesystem fs.FS
	mappings   map[string]*VersionedMapping
}

func NewMappingsRegistry(config *Config, filesystem fs.FS) MappingsRegistry {
	return &mappingsRegistry{
		config:     config,
		filesystem: filesystem,
		mappings:   make(map[string]*VersionedMapping),
	}
}

func (mr *mappingsRegistry) LoadMappings() error {
	mappingsDir := mr.config.MappingsPath
	files, err := fs.ReadDir(mr.filesystem, mappingsDir)
	if err != nil {
		return fmt.Errorf(`error finding mappings in directory: %s`, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		documentKind := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		versionedMappingJson, err := fs.ReadFile(mr.filesystem, filepath.Join(mappingsDir, file.Name()))

		if err != nil {
			return fmt.Errorf(`error reading file: %s`, err)
		}

		mapping := &VersionedMapping{}
		if err := json.Unmarshal(versionedMappingJson, mapping); err != nil {
			return fmt.Errorf(`invalid json in file "%s": %s`, file.Name(), err)
		}

		mr.mappings[documentKind] = mapping
	}

	return nil
}

func (mr *mappingsRegistry) ParseIndexName(indexName string) *IndexName {
	// the index name is assumed to match one of the following types
	// the documentKind may contain the delimiter
	// indexPrefix-version-documentKind
	// indexPrefix-version-innerName-documentKind
	indexName = strings.TrimPrefix(indexName, mr.config.IndexPrefix+indexNamePartsDelimiter)
	documentKind := ""

	for k := range mr.mappings {
		if strings.HasSuffix(indexName, indexNamePartsDelimiter+k) {
			documentKind = k
			indexName = strings.TrimSuffix(indexName, indexNamePartsDelimiter+documentKind)
			break
		}
	}

	if documentKind == "" {
		return nil
	}

	version := mr.Version(documentKind)
	indexName = strings.TrimPrefix(indexName, version)

	return &IndexName{
		DocumentKind: documentKind,
		Version:      version,
		Inner:        strings.TrimPrefix(indexName, indexNamePartsDelimiter),
	}
}

func (mr *mappingsRegistry) IndexName(documentKind, inner string) string {
	return nonEmptyJoin([]string{
		mr.config.IndexPrefix,
		mr.Version(documentKind),
		inner,
		documentKind,
	}, indexNamePartsDelimiter)
}

func (mr *mappingsRegistry) AliasName(documentKind, inner string) string {
	return nonEmptyJoin([]string{
		mr.config.IndexPrefix,
		inner,
		documentKind,
	}, indexNamePartsDelimiter)
}

func (mr *mappingsRegistry) Version(documentKind string) string {
	mapping := mr.Mapping(documentKind)
	if mapping == nil {
		return ""
	}

	return mapping.Version
}

func (mr *mappingsRegistry) Mapping(documentKind string) *VersionedMapping {
	mapping, ok := mr.mappings[documentKind]
	if !ok {
		return nil
	}

	return mapping
}

func nonEmptyJoin(parts []string, delimiter string) string {
	var nonEmpty []string
	for _, str := range parts {
		if str != "" {
			nonEmpty = append(nonEmpty, str)
		}
	}

	return strings.Join(nonEmpty, delimiter)
}

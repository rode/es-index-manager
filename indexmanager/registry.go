package indexmanager

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
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
}

type mappingsRegistry struct {
	config   *Config
	filesystem fs.FS
	mappings map[string]*VersionedMapping
}

func NewMappingsRegistry(config *Config, filesystem fs.FS) MappingsRegistry {
	return &mappingsRegistry{
		config:   config,
		mappings: make(map[string]*VersionedMapping),
	}
}

func (mr *mappingsRegistry) LoadMappings() error {
	mappingsDir := mr.config.MappingsPath
	files, err := fs.ReadDir(mr.filesystem, mappingsDir)
	if err != nil {
		return fmt.Errorf(`error finding mappings in directory "%s": %s`, mappingsDir, err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to determine working directory: %s", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		documentKind := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		filePath := path.Join(currentDir, mappingsDir, file.Name())
		versionedMappingJson, err := fs.ReadFile(mr.filesystem, filePath)

		if err != nil {
			return fmt.Errorf(`error reading file "%s": %s`, filePath, err)
		}

		var mapping *VersionedMapping
		if err := json.Unmarshal(versionedMappingJson, mapping); err != nil {
			return fmt.Errorf(`invalid json in file "%s": %s`, filePath, err)
		}

		mr.mappings[documentKind] = mapping
	}

	return nil
}

func (mr *mappingsRegistry) IndexName(documentKind, inner string) string {
	return strings.Join([]string{
		mr.config.IndexPrefix,
		mr.Version(documentKind),
		inner,
		documentKind,
	}, indexNamePartsDelimiter)
}

func (mr *mappingsRegistry) AliasName(documentKind, inner string) string {
	return strings.Join([]string{
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

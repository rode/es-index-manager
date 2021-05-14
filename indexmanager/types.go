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

import "time"

const (
	indexNamePartsDelimiter = "-"
)

type Migration struct {
	Alias        string
	SourceIndex  string
	TargetIndex  string
	DocumentKind string
}

type MigrationConfig struct {
	PollInterval time.Duration
	PollAttempts int
}

type Config struct {
	IndexPrefix  string
	MappingsPath string
	Migration    *MigrationConfig
}

type VersionedMapping struct {
	Version  string                 `json:"version"`
	Mappings map[string]interface{} `json:"mappings"`
}

type IndexName struct {
	DocumentKind string
	Version      string
	Inner        string // the parts that aren't the prefix, document kind, or version
}

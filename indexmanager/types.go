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
	// PollInterval is the time to wait between polls of the reindex task endpoint.
	PollInterval time.Duration
	// PollAttempts is the number of times that the IndexManager will fetch the task document
	// to check if the reindex has finished.
	PollAttempts int
}

type Config struct {
	// IndexPrefix is used when creating index and alias names, and to tell if a particular index is associated with the
	// application. The IndexManager only operates on indices with this prefix; in addition, any indices must have the
	// prefix be the value of the _meta.type on the index mapping.
	IndexPrefix string
	// MappingsPath should point to a directory containing JSON files that hold the Elasticsearch mappings for a given
	// document kind. The document kind is assumed to be the name of the file, after removing the extension. The version
	// and mappings are top level keys:
	//
	// {
	//  "version": "v1alpha1",
	//  "mappings": {
	//    "_meta": {
	//      "type": "myApp"
	//    }
	//  }
	//}
	MappingsPath string
	// Migration controls the amount of time the IndexManager will wait for a reindex to complete as part of a migration.
	Migration *MigrationConfig
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

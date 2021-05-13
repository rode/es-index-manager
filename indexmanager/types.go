package indexmanager

import "time"

// IndexManager

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

// Elasticsearch

const (
	elasticSearchAllIndices            = "_all"
	elasticsearchTaskIndex             = ".tasks"
	elasticsearchResourceAlreadyExists = "resource_already_exists_exception"
)

// TODO: fix this structure
// Elasticsearch 400 response
type esErrorResponse struct {
	Error esError `json:"error"`
}

type esError struct {
	Type string `json:"type"`
}

type esIndex struct {
	Mappings *esMappings `json:"mappings"`
}

type esMappings struct {
	Meta *esMeta `json:"_meta,omitempty"`
}

type esMeta struct {
	Type string `json:"type,omitempty"`
}

// Elasticsearch /$INDEX/block/_write response

type esBlockIndex struct {
	Name    string `json:"name"`
	Blocked bool   `json:"blocked"`
}

type esBlockResponse struct {
	Acknowledged       bool           `json:"acknowledged"`
	ShardsAcknowledged bool           `json:"shards_acknowledged"`
	Indices            []esBlockIndex `json:"indices"`
}

// Elasticsearch /$INDEX/_settings
// TODO: fix this structure
type esSettingsResponse struct {
	Settings *ESSettingsIndex `json:"settings"`
}

type ESSettingsIndex struct {
	Index *esSettingsBlocks `json:"index"`
}

type esSettingsBlocks struct {
	Blocks *esSettingsWrite `json:"blocks"`
}

type esSettingsWrite struct {
	Write string `json:"write"`
}

// response for calls where wait_for_completion=false
type esTaskCreationResponse struct {
	Task string `json:"task"`
}

// /_tasks/$TASK_ID response
type esTask struct {
	Completed bool `json:"completed"`
}

// Elasticsearch /_aliases request
type esActions struct {
	Add    *esIndexAlias `json:"add,omitempty"`
	Remove *esIndexAlias `json:"remove,omitempty"`
}

type esIndexAlias struct {
	Index string `json:"index"`
	Alias string `json:"alias"`
}

type esIndexAliasRequest struct {
	Actions []esActions `json:"actions"`
}

// Elasticsearch /_reindex request

type esReindex struct {
	Conflicts   string           `json:"conflicts"`
	Source      *esReindexFields `json:"source"`
	Destination *esReindexFields `json:"dest"`
}

type esReindexFields struct {
	Index  string `json:"index"`
	OpType string `json:"op_type,omitempty"`
}

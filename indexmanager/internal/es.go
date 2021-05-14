package internal

// Elasticsearch
const (
	ElasticsearchAllIndices            = "_all"
	ElasticsearchTaskIndex             = ".tasks"
	ElasticsearchResourceAlreadyExists = "resource_already_exists_exception"
)

// Elasticsearch 400 response
type EsErrorResponse struct {
	Error EsError `json:"error"`
}

type EsError struct {
	Type string `json:"type"`
}

type EsIndex struct {
	Mappings *EsMappings `json:"mappings"`
}

type EsMappings struct {
	Meta *EsMeta `json:"_meta,omitempty"`
}

type EsMeta struct {
	Type string `json:"type,omitempty"`
}

// Elasticsearch /$INDEX/block/_write response

type EsBlockIndex struct {
	Name    string `json:"name"`
	Blocked bool   `json:"blocked"`
}

type EsBlockResponse struct {
	Acknowledged       bool           `json:"acknowledged"`
	ShardsAcknowledged bool           `json:"shards_acknowledged"`
	Indices            []EsBlockIndex `json:"indices"`
}

// Elasticsearch /$INDEX/_settings
type EsSettingsResponse struct {
	Settings *EsSettingsIndex `json:"settings"`
}

type EsSettingsIndex struct {
	Index *EsSettingsBlocks `json:"index"`
}

type EsSettingsBlocks struct {
	Blocks *EsSettingsWrite `json:"blocks"`
}

type EsSettingsWrite struct {
	Write string `json:"write"`
}

// response for calls where wait_for_completion=false
type EsTaskCreationResponse struct {
	Task string `json:"task"`
}

// /_tasks/$TASK_ID response
type EsTask struct {
	Completed bool `json:"completed"`
}

// Elasticsearch /_aliases request
type EsActions struct {
	Add    *EsIndexAlias `json:"add,omitempty"`
	Remove *EsIndexAlias `json:"remove,omitempty"`
}

type EsIndexAlias struct {
	Index string `json:"index"`
	Alias string `json:"alias"`
}

type EsIndexAliasRequest struct {
	Actions []EsActions `json:"actions"`
}

// Elasticsearch /_reindex request

type EsReindex struct {
	Conflicts   string           `json:"conflicts"`
	Source      *EsReindexFields `json:"source"`
	Destination *EsReindexFields `json:"dest"`
}

type EsReindexFields struct {
	Index  string `json:"index"`
	OpType string `json:"op_type,omitempty"`
}

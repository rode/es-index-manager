# es-index-manager

A Go library for creating Elasticsearch indices and applying mapping changes.

## Assumptions

The `IndexManager` makes a couple of assumptions about how the consuming application uses
Elasticsearch indices:

- Indices are assumed to start with a common prefix. They should also include that prefix as the value of the key `_meta.type` in
the index mappings. This is to prevent the `IndexManager` from trying to operate on indices that are internal to Elasticsearch or that are owned by other applications.
The value of this prefix can be specified in `Config.IndexPrefix`.
- An index only stores a single type (or kind) of document.
- It's acceptable for there to be a window of time during which writes are disallowed, while the `IndexManager` is applying the latest mappings
to a new index and reindexing.

## Use

```go
package main

import (
	"context"
	"time"

	"github.com/rode/es-index-manager/indexmanager"
	"go.uber.org/zap"
)

func main() {
	config := &indexmanager.Config{
		MappingsPath: "mappings",
		IndexPrefix:  "myapp",
		// the migration config can be omitted and will default to 10 attempts with an interval of 10 seconds
		Migration: &indexmanager.MigrationConfig{
			PollInterval: 30 * time.Second,
			PollAttempts: 5,
		},
	}
	logger, _ := zap.Development()
	client, _ := elasticsearch.NewClient(elasticsearch.Config{})

	manager := indexmanager.NewIndexManager(logger.Named("IndexManager"), client, config)
	// load the mappings from Config.MappingsPath and run any needed migrations
	manager.Initialize(context.Background())
	// creates a new index called myapp-v1-foo-bar, where bar is the document kind
	manager.CreateIndex(context.Background(), manager.IndexName("bar", "foo"), "", "bar")
}
```

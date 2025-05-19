package logger

import (
	"context"
	"time"

	"github.com/ncobase/ncore/data/search/opensearch"
	"github.com/sirupsen/logrus"
)

// OpenSearchHook represents an OpenSearch log hook
type OpenSearchHook struct {
	client *opensearch.Client
	index  string
}

// Levels returns all log levels
func (h *OpenSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends log entry to OpenSearch
func (h *OpenSearchHook) Fire(entry *logrus.Entry) error {
	// Generate document ID from timestamp
	docID := entry.Time.Format(time.RFC3339Nano)

	// Create context for the operation
	ctx := context.Background()

	// Index the log entry
	return h.client.IndexDocument(
		ctx,
		h.index,
		docID,
		entry.Data,
	)
}

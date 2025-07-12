package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ncobase/ncore/data/search/opensearch"
	"github.com/ncobase/ncore/logging/logger/config"
	"github.com/sirupsen/logrus"
)

// OpenSearchHook represents an OpenSearch log hook
type OpenSearchHook struct {
	client *opensearch.Client
	config *config.Config
}

// NewOpenSearchHook creates new OpenSearch hook
func NewOpenSearchHook(client *opensearch.Client, cfg *config.Config) *OpenSearchHook {
	return &OpenSearchHook{
		client: client,
		config: cfg,
	}
}

// Levels returns all log levels
func (h *OpenSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends log entry to OpenSearch
func (h *OpenSearchHook) Fire(entry *logrus.Entry) error {
	// Get current index name with date suffix
	currentIndex := h.getCurrentIndexName(entry.Time)

	// Prepare log document
	logDoc := h.prepareLogDocument(entry)

	// Generate unique document ID based on timestamp and nanoseconds
	docID := fmt.Sprintf("%d-%d", entry.Time.UnixNano(), time.Now().Nanosecond())

	// Index the log entry
	return h.client.IndexDocument(
		context.Background(),
		currentIndex,
		docID,
		logDoc,
	)
}

// getCurrentIndexName returns index name
func (h *OpenSearchHook) getCurrentIndexName(logTime time.Time) string {
	if h.config == nil {
		return "default-log"
	}
	return h.config.BuildIndexName(logTime)
}

// prepareLogDocument prepares the log document structure
func (h *OpenSearchHook) prepareLogDocument(entry *logrus.Entry) map[string]any {
	logDoc := make(map[string]any)

	// Add timestamp in multiple formats for compatibility
	logDoc["@timestamp"] = entry.Time.Format(time.RFC3339)
	logDoc["timestamp"] = entry.Time.UnixMilli()
	logDoc["time"] = entry.Time.Format(time.RFC3339)

	// Add log level and message
	logDoc["level"] = entry.Level.String()
	logDoc["message"] = entry.Message

	// Add hostname if available
	if hostname, err := os.Hostname(); err == nil {
		logDoc["hostname"] = hostname
	}

	// Add all fields from entry
	for key, value := range entry.Data {
		// Avoid overwriting system fields
		if key != "@timestamp" && key != "timestamp" && key != "level" && key != "message" {
			logDoc[key] = value
		}
	}

	return logDoc
}

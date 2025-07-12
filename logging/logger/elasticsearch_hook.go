package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ncobase/ncore/data/search/elastic"
	"github.com/ncobase/ncore/logging/logger/config"
	"github.com/sirupsen/logrus"
)

// ElasticSearchHook represents an Elasticsearch log hook
type ElasticSearchHook struct {
	client *elastic.Client
	config *config.Config
}

// NewElasticSearchHook creates new Elasticsearch hook
func NewElasticSearchHook(client *elastic.Client, cfg *config.Config) *ElasticSearchHook {
	return &ElasticSearchHook{
		client: client,
		config: cfg,
	}
}

// Levels returns all log levels
func (h *ElasticSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends log entry to Elasticsearch
func (h *ElasticSearchHook) Fire(entry *logrus.Entry) error {
	// Get current index name with date suffix
	currentIndex := h.getCurrentIndexName(entry.Time)

	// Prepare log document
	logDoc := h.prepareLogDocument(entry)

	// Check if index might be a data stream
	if h.isDataStreamIndex(currentIndex) {
		return h.indexAsDataStream(currentIndex, logDoc)
	}

	// Use regular index operation
	return h.client.IndexDocument(
		context.Background(),
		currentIndex,
		entry.Time.Format(timeFormat),
		logDoc,
	)
}

// getCurrentIndexName returns index name
func (h *ElasticSearchHook) getCurrentIndexName(logTime time.Time) string {
	if h.config == nil {
		return "default-log"
	}
	return h.config.BuildIndexName(logTime)
}

// prepareLogDocument prepares the log document structure
func (h *ElasticSearchHook) prepareLogDocument(entry *logrus.Entry) map[string]any {
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

// isDataStreamIndex checks if index name might be treated as data stream by ES
func (h *ElasticSearchHook) isDataStreamIndex(indexName string) bool {
	lowerIndex := strings.ToLower(indexName)
	dataStreamPrefixes := []string{"logs-", "metrics-", "traces-"}
	for _, prefix := range dataStreamPrefixes {
		if strings.HasPrefix(lowerIndex, prefix) {
			return true
		}
	}
	return false
}

// indexAsDataStream handles indexing for data streams
func (h *ElasticSearchHook) indexAsDataStream(indexName string, logDoc map[string]any) error {
	client := h.client.GetClient()
	if client == nil {
		return fmt.Errorf("elasticsearch client is nil")
	}

	// Convert document to JSON
	docBytes, err := json.Marshal(logDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal log document: %w", err)
	}

	// Use Index operation with op_type=create for data streams
	res, err := client.Index(
		indexName,
		strings.NewReader(string(docBytes)),
		client.Index.WithOpType("create"),
		client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("elasticsearch index error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch index error: %s", res.Status())
	}

	return nil
}

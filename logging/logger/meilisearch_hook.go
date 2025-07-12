package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger/config"
	"github.com/sirupsen/logrus"
)

// MeiliSearchHook represents a MeiliSearch log hook
type MeiliSearchHook struct {
	client *meili.Client
	config *config.Config
}

// NewMeiliSearchHook creates new MeiliSearch hook
func NewMeiliSearchHook(client *meili.Client, cfg *config.Config) *MeiliSearchHook {
	return &MeiliSearchHook{
		client: client,
		config: cfg,
	}
}

// Levels returns all log levels
func (h *MeiliSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends log entry to MeiliSearch
func (h *MeiliSearchHook) Fire(entry *logrus.Entry) error {
	// Get current index name with date suffix
	currentIndex := h.getCurrentIndexName(entry.Time)

	// Prepare log document
	logDoc := h.prepareLogDocument(entry)

	// Index as array with single document
	return h.client.IndexDocuments(currentIndex, []any{logDoc})
}

// getCurrentIndexName returns index name
func (h *MeiliSearchHook) getCurrentIndexName(logTime time.Time) string {
	if h.config == nil {
		return "default-log"
	}
	return h.config.BuildIndexName(logTime)
}

// prepareLogDocument prepares the log document structure for MeiliSearch
func (h *MeiliSearchHook) prepareLogDocument(entry *logrus.Entry) map[string]any {
	logDoc := make(map[string]any)

	// Add unique ID for MeiliSearch (required)
	logDoc["id"] = fmt.Sprintf("%d-%d", entry.Time.UnixNano(), time.Now().Nanosecond())

	// Add timestamp fields
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
		if key != "@timestamp" && key != "timestamp" && key != "level" && key != "message" && key != "id" {
			logDoc[key] = value
		}
	}

	return logDoc
}

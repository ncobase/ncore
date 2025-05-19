package logger

import (
	"encoding/json"
	"fmt"

	"github.com/ncobase/ncore/data/search/meili"
	"github.com/sirupsen/logrus"
)

// MeiliSearchHook represents a MeiliSearch log hook
type MeiliSearchHook struct {
	client *meili.Client
	index  string
}

// Levels returns all log levels
func (h *MeiliSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends log entry to MeiliSearch
func (h *MeiliSearchHook) Fire(entry *logrus.Entry) error {
	jsonData, err := json.Marshal(entry.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal log data: %w", err)
	}
	return h.client.IndexDocuments(h.index, jsonData)
}

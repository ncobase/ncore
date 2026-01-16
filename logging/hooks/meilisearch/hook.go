// Package meilisearch provides a logrus hook for sending logs to Meilisearch.
package meilisearch

import (
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/logging/logger/config"
	"github.com/sirupsen/logrus"
)

func init() {
	logger.RegisterHookFactory(logger.HookMeilisearch, NewHook)
}

// Hook is a logrus hook for Meilisearch
type Hook struct {
	client      meilisearch.ServiceManager
	indexName   string
	dateSuffix  string
	rotateDaily bool
	levels      []logrus.Level
}

// NewHook creates a new Meilisearch hook from config
func NewHook(cfg *config.Config) (logrus.Hook, error) {
	if cfg.Meilisearch == nil {
		return nil, fmt.Errorf("meilisearch config is nil")
	}

	client := meilisearch.New(cfg.Meilisearch.Host, meilisearch.WithAPIKey(cfg.Meilisearch.APIKey))

	// Test connection
	if _, err := client.Health(); err != nil {
		return nil, fmt.Errorf("failed to connect to meilisearch: %w", err)
	}

	return &Hook{
		client:      client,
		indexName:   cfg.IndexName,
		dateSuffix:  cfg.DateSuffix,
		rotateDaily: cfg.RotateDaily,
		levels:      logrus.AllLevels,
	}, nil
}

// Fire sends the log entry to Meilisearch
func (h *Hook) Fire(entry *logrus.Entry) error {
	indexName := h.buildIndexName()

	doc := map[string]any{
		"id":        fmt.Sprintf("%d", time.Now().UnixNano()),
		"timestamp": entry.Time.UTC().Format(time.RFC3339Nano),
		"level":     entry.Level.String(),
		"message":   entry.Message,
	}

	for k, v := range entry.Data {
		doc[k] = v
	}

	index := h.client.Index(indexName)
	pk := "id"
	_, err := index.AddDocuments([]map[string]any{doc}, &meilisearch.DocumentOptions{PrimaryKey: &pk})
	if err != nil {
		return fmt.Errorf("failed to index log entry: %w", err)
	}

	return nil
}

// Levels returns the log levels this hook fires for
func (h *Hook) Levels() []logrus.Level {
	return h.levels
}

func (h *Hook) buildIndexName() string {
	if !h.rotateDaily {
		return h.indexName
	}
	dateSuffix := time.Now().Format(h.dateSuffix)
	return fmt.Sprintf("%s-%s", h.indexName, dateSuffix)
}

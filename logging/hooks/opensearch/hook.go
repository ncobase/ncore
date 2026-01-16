// Package opensearch provides a logrus hook for sending logs to OpenSearch.
package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/logging/logger/config"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/sirupsen/logrus"
)

func init() {
	logger.RegisterHookFactory(logger.HookOpenSearch, NewHook)
}

// Hook is a logrus hook for OpenSearch
type Hook struct {
	client      *opensearchapi.Client
	indexName   string
	dateSuffix  string
	rotateDaily bool
	levels      []logrus.Level
}

// NewHook creates a new OpenSearch hook from config
func NewHook(cfg *config.Config) (logrus.Hook, error) {
	if cfg.OpenSearch == nil {
		return nil, fmt.Errorf("opensearch config is nil")
	}

	client, err := opensearchapi.NewClient(
		opensearchapi.Config{
			Client: opensearch.Config{
				Addresses: cfg.OpenSearch.Addresses,
				Username:  cfg.OpenSearch.Username,
				Password:  cfg.OpenSearch.Password,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create opensearch client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Info(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to opensearch: %w", err)
	}

	return &Hook{
		client:      client,
		indexName:   cfg.IndexName,
		dateSuffix:  cfg.DateSuffix,
		rotateDaily: cfg.RotateDaily,
		levels:      logrus.AllLevels,
	}, nil
}

// Fire sends the log entry to OpenSearch
func (h *Hook) Fire(entry *logrus.Entry) error {
	indexName := h.buildIndexName()

	doc := map[string]any{
		"@timestamp": entry.Time.UTC().Format(time.RFC3339Nano),
		"level":      entry.Level.String(),
		"message":    entry.Message,
	}

	for k, v := range entry.Data {
		doc[k] = v
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := opensearchapi.IndexReq{
		Index: indexName,
		Body:  strings.NewReader(string(body)),
	}

	_, err = h.client.Index(ctx, req)
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

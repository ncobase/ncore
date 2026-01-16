// Package elasticsearch provides a logrus hook for sending logs to Elasticsearch.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/logging/logger/config"
	"github.com/sirupsen/logrus"
)

func init() {
	logger.RegisterHookFactory(logger.HookElasticsearch, NewHook)
}

// Hook is a logrus hook for Elasticsearch
type Hook struct {
	client      *elasticsearch.Client
	indexName   string
	dateSuffix  string
	rotateDaily bool
	levels      []logrus.Level
}

// NewHook creates a new Elasticsearch hook from config
func NewHook(cfg *config.Config) (logrus.Hook, error) {
	if cfg.Elasticsearch == nil {
		return nil, fmt.Errorf("elasticsearch config is nil")
	}

	esCfg := elasticsearch.Config{
		Addresses: cfg.Elasticsearch.Addresses,
	}
	if cfg.Elasticsearch.Username != "" {
		esCfg.Username = cfg.Elasticsearch.Username
		esCfg.Password = cfg.Elasticsearch.Password
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch connection error: %s", res.Status())
	}

	return &Hook{
		client:      client,
		indexName:   cfg.IndexName,
		dateSuffix:  cfg.DateSuffix,
		rotateDaily: cfg.RotateDaily,
		levels:      logrus.AllLevels,
	}, nil
}

// Fire sends the log entry to Elasticsearch
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

	res, err := h.client.Index(
		indexName,
		bytes.NewReader(body),
		h.client.Index.WithContext(ctx),
		h.client.Index.WithRefresh("false"),
	)
	if err != nil {
		return fmt.Errorf("failed to index log entry: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch index error: %s", res.Status())
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

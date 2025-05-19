package logger

import (
	"context"

	"github.com/ncobase/ncore/data/search/elastic"
	"github.com/sirupsen/logrus"
)

// ElasticSearchHook represents an Elasticsearch log hook
type ElasticSearchHook struct {
	client *elastic.Client
	index  string
}

// Levels returns all log levels
func (h *ElasticSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends log entry to Elasticsearch
func (h *ElasticSearchHook) Fire(entry *logrus.Entry) error {
	return h.client.IndexDocument(
		context.Background(),
		h.index,
		entry.Time.Format(timeFormat),
		entry.Data,
	)
}

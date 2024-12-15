package connection

import (
	"context"
	"errors"
	"ncobase/common/config"
	"ncobase/common/data/meili"
	"ncobase/common/logger"
)

// newMeilisearchClient creates a new Meilisearch client
func newMeilisearchClient(conf *config.Meilisearch) (*meili.Client, error) {
	if conf == nil || conf.Host == "" {
		logger.Infof(context.Background(), "Meilisearch configuration is nil or empty")
		return nil, errors.New("meilisearch configuration is nil or empty")
	}

	ms := meili.NewMeilisearch(conf.Host, conf.APIKey)

	if _, err := ms.GetClient().Health(); err != nil {
		logger.Errorf(context.Background(), "Meilisearch connect error: %v", err)
		return nil, err
	}

	logger.Infof(context.Background(), "Meilisearch connected")

	return ms, nil
}

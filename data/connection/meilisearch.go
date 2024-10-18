package connection

import (
	"context"
	"errors"
	"ncobase/common/config"
	"ncobase/common/log"
	"ncobase/common/meili"
)

// newMeilisearchClient creates a new Meilisearch client
func newMeilisearchClient(conf *config.Meilisearch) (*meili.Client, error) {
	if conf == nil || conf.Host == "" {
		log.Infof(context.Background(), "Meilisearch configuration is nil or empty")
		return nil, errors.New("meilisearch configuration is nil or empty")
	}

	ms := meili.NewMeilisearch(conf.Host, conf.APIKey)

	if _, err := ms.GetClient().Health(); err != nil {
		log.Errorf(context.Background(), "Meilisearch connect error: %v", err)
		return nil, err
	}

	log.Infof(context.Background(), "Meilisearch connected")

	return ms, nil
}

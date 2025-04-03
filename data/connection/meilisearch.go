package connection

import (
	"fmt"
	"ncobase/ncore/data/config"
	"ncobase/ncore/data/meili"
)

// newMeilisearchClient creates a new Meilisearch client
func newMeilisearchClient(conf *config.Meilisearch) (*meili.Client, error) {
	if conf == nil || conf.Host == "" {
		return nil, fmt.Errorf("meilisearch configuration is nil or empty")
	}

	ms := meili.NewMeilisearch(conf.Host, conf.APIKey)

	if _, err := ms.GetClient().Health(); err != nil {
		return nil, fmt.Errorf("meilisearch connect error: %v", err)
	}

	return ms, nil
}

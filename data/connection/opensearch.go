package connection

import (
	"context"
	"errors"
	"fmt"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/search/opensearch"
)

// newOpenSearchClient creates a new OpenSearch client
func newOpenSearchClient(conf *config.OpenSearch) (*opensearch.Client, error) {
	if conf == nil || len(conf.Addresses) == 0 {
		return nil, errors.New("opensearch configuration is nil or empty")
	}

	os, err := opensearch.NewClient(conf.Addresses, conf.Username, conf.Password, conf.InsecureSkipTLS)
	if err != nil {
		return nil, fmt.Errorf("opensearch client creation error: %w", err)
	}

	// Test connection
	health, err := os.Health(context.Background())
	if err != nil {
		return nil, fmt.Errorf("opensearch connect error: %w", err)
	}

	// Log cluster health status
	fmt.Printf("OpenSearch cluster status: %s\n", health)

	return os, nil
}

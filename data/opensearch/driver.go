// Package opensearch provides an OpenSearch driver for ncore/data.
//
// This driver uses opensearch-go (github.com/opensearch-project/opensearch-go/v2)
// as the underlying client. It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/opensearch"
//
// OpenSearch is AWS's fork of Elasticsearch, providing full compatibility with
// Elasticsearch APIs while adding additional features.
//
// Example usage:
//
//	driver, err := data.GetSearchDriver("opensearch")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cfg := &config.OpenSearch{
//	    Addresses: []string{"https://localhost:9200"},
//	    Username:  "admin",
//	    Password:  "admin",
//	}
//
//	conn, err := driver.Connect(ctx, cfg)
//	client := conn.(*client.Client)
package opensearch

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/opensearch/client"
)

// driver implements data.SearchDriver for OpenSearch.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "opensearch"
}

// Connect establishes an OpenSearch connection using the provided configuration.
//
// The configuration must be a *config.OpenSearch containing:
//   - Addresses: List of OpenSearch node URLs
//   - Username: Optional authentication username (default: "admin")
//   - Password: Optional authentication password
//
// Example addresses:
//
//	[]string{"https://localhost:9200"}
//	[]string{"https://search-domain.us-east-1.es.amazonaws.com"}
//
// Returns an *client.Client wrapper that provides search, indexing, and
// document management operations compatible with Elasticsearch APIs.
func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	osCfg, ok := cfg.(*config.OpenSearch)
	if !ok {
		return nil, fmt.Errorf("opensearch: invalid configuration type, expected *config.OpenSearch")
	}

	if len(osCfg.Addresses) == 0 {
		return nil, fmt.Errorf("opensearch: addresses are empty")
	}

	client, err := client.NewClient(osCfg.Addresses, osCfg.Username, osCfg.Password, false)
	if err != nil {
		return nil, fmt.Errorf("opensearch: failed to create client: %w", err)
	}

	return client, nil
}

// Close terminates the OpenSearch connection and releases resources.
func (d *driver) Close(conn any) error {
	_, ok := conn.(*client.Client)
	if !ok {
		return fmt.Errorf("opensearch: invalid connection type, expected *client.Client")
	}

	return nil
}

// init registers the OpenSearch driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterSearchDriver(&driver{})
}

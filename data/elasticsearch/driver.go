// Package elasticsearch provides an Elasticsearch driver for ncore/data.
//
// This driver uses go-elasticsearch/v8 (github.com/elastic/go-elasticsearch/v8) as
// the underlying client. It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/elasticsearch"
//
// The driver supports Elasticsearch connection configuration including cluster
// addresses, authentication, and health checking.
//
// Example usage:
//
//	driver, err := data.GetSearchDriver("elasticsearch")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cfg := &config.Elasticsearch{
//	    Addresses: []string{"http://localhost:9200"},
//	    Username:  "elastic",
//	    Password:  "password",
//	}
//
//	conn, err := driver.Connect(ctx, cfg)
//	client := conn.(*elastic.Client)
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/elasticsearch/client"
)

// driver implements data.SearchDriver for Elasticsearch.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "elasticsearch"
}

// Connect establishes an Elasticsearch connection using the provided configuration.
//
// The configuration must be a *config.Elasticsearch containing:
//   - Addresses: List of Elasticsearch node URLs
//   - Username: Optional authentication username
//   - Password: Optional authentication password
//
// Example addresses:
//
//	[]string{"http://localhost:9200"}
//	[]string{"https://es1.example.com:9200", "https://es2.example.com:9200"}
//
// Returns an *elastic.Client wrapper that provides search, indexing, and
// document management operations.
func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	esCfg, ok := cfg.(*config.Elasticsearch)
	if !ok {
		return nil, fmt.Errorf("elasticsearch: invalid configuration type, expected *config.Elasticsearch")
	}

	if len(esCfg.Addresses) == 0 {
		return nil, fmt.Errorf("elasticsearch: addresses are empty")
	}

	client, err := client.NewClient(esCfg.Addresses, esCfg.Username, esCfg.Password)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch: failed to create client: %w", err)
	}

	return client, nil
}

// Close terminates the Elasticsearch connection and releases resources.
func (d *driver) Close(conn any) error {
	_, ok := conn.(*client.Client)
	if !ok {
		return fmt.Errorf("elasticsearch: invalid connection type, expected *client.Client")
	}

	return nil
}

// init registers the Elasticsearch driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterSearchDriver(&driver{})
}

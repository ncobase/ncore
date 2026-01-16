// Package meilisearch provides a Meilisearch driver for ncore/data.
//
// This driver uses meilisearch-go (github.com/meilisearch/meilisearch-go) as the underlying client.
// It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/meilisearch"
//
// The driver supports Meilisearch connection configuration including host and API key.
package meilisearch

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/meilisearch/client"
)

// driver implements data.SearchDriver for Meilisearch.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "meilisearch"
}

// Connect establishes a Meilisearch connection using the provided configuration.
//
// The configuration must be a *config.Meilisearch containing:
//   - Host: Meilisearch server address (e.g., "http://localhost:7700")
//   - APIKey: API key for authentication (optional for development)
//
// Example configuration:
//
//	&config.Meilisearch{
//	    Host:   "http://localhost:7700",
//	    APIKey: "masterKey",
//	}
//
// The connection is verified with a health check before being returned.
// Returns a *meili.Client on success.
func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	msCfg, ok := cfg.(*config.Meilisearch)
	if !ok {
		return nil, fmt.Errorf("meilisearch: invalid configuration type, expected *config.Meilisearch")
	}

	if msCfg.Host == "" {
		return nil, fmt.Errorf("meilisearch: host is empty")
	}

	// Create Meilisearch client using the existing wrapper
	client := client.NewMeilisearch(msCfg.Host, msCfg.APIKey)

	// Verify the connection works by checking health
	if _, err := client.Health(); err != nil {
		return nil, fmt.Errorf("meilisearch: health check failed: %w", err)
	}

	return client, nil
}

// Close terminates the Meilisearch connection and releases resources.
//
// Note: The meilisearch-go SDK does not require explicit connection cleanup,
// but this method is provided for interface compliance.
func (d *driver) Close(conn any) error {
	_, ok := conn.(*client.Client)
	if !ok {
		return fmt.Errorf("meilisearch: invalid connection type, expected *meilisearch.Client")
	}

	// The meilisearch-go client uses HTTP and doesn't require explicit cleanup
	// This is a no-op but maintains interface compliance
	return nil
}

// init registers the Meilisearch driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterSearchDriver(&driver{})
}

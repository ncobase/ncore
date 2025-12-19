package data

import (
	"database/sql"
	"errors"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/search"
	"github.com/ncobase/ncore/data/search/elastic"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/data/search/opensearch"
	"github.com/redis/go-redis/v9"
)

// Database Access Methods

// GetDBManager returns the database manager
func (d *Data) GetDBManager() *connection.DBManager {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.DBM
}

// GetMasterDB returns the master database connection for write operations
func (d *Data) GetMasterDB() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.DB()
}

// GetSlaveDB returns slave database connection for read operations
func (d *Data) GetSlaveDB() (*sql.DB, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return nil, errors.New("data layer is closed")
	}
	if d.Conn == nil {
		return nil, errors.New("no database connection available")
	}
	return d.Conn.ReadDB()
}

// Cache Access Methods

// GetRedis returns the Redis client
func (d *Data) GetRedis() *redis.Client {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}

	// Update connection metrics if client exists
	if d.Conn.RC != nil {
		d.collector.RedisConnections(int(d.Conn.RC.PoolStats().TotalConns))
	}
	return d.Conn.RC
}

// MongoDB Access Methods

// GetMongoManager returns the MongoDB manager
func (d *Data) GetMongoManager() *connection.MongoManager {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.MGM
}

// Search Engine Access Methods

// initSearchClient initializes search client lazily and safely
func (d *Data) initSearchClient() {
	d.searchOnce.Do(func() {
		d.mu.RLock()
		if d.searchClient != nil {
			d.mu.RUnlock()
			return
		}
		d.mu.RUnlock()

		d.mu.Lock()
		defer d.mu.Unlock()

		if d.searchClient != nil {
			return
		}

		if !d.closed {
			var searchConfig *config.Search
			if d.conf != nil {
				searchConfig = d.conf.Search
			}

			d.searchClient = search.NewClientWithConfig(
				d.getElasticsearchUnsafe(),
				d.getOpenSearchUnsafe(),
				d.getMeilisearchUnsafe(),
				d.collector,
				searchConfig,
			)
		}
	})
}

// SetSearchIndexPrefix sets custom index prefix for search operations
func (d *Data) SetSearchIndexPrefix(prefix string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conf != nil && d.conf.Search != nil {
		d.conf.Search.IndexPrefix = prefix
	}

	if d.searchClient != nil {
		d.searchClient.SetIndexPrefix(prefix)
	}
}

// GetSearchIndexPrefix returns current search index prefix
func (d *Data) GetSearchIndexPrefix() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.searchClient != nil {
		return d.searchClient.GetIndexPrefix()
	}

	if d.conf != nil && d.conf.Search != nil {
		return d.conf.Search.IndexPrefix
	}

	return ""
}

// GetSearchConfig returns current search configuration
func (d *Data) GetSearchConfig() *config.Search {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.conf != nil {
		return d.conf.Search
	}
	return nil
}

// UpdateSearchConfig updates search configuration
func (d *Data) UpdateSearchConfig(searchConfig *config.Search) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conf != nil {
		d.conf.Search = searchConfig
	}

	if d.searchClient != nil {
		d.searchClient.UpdateSearchConfig(searchConfig)
	}
}

// GetSearchClient returns initialized search client that abstracts Meilisearch, Elasticsearch, and OpenSearch.
// The client automatically selects the best available search engine based on configuration.
// Returns nil if no search engine is configured.
func (d *Data) GetSearchClient() *search.Client {
	d.initSearchClient()

	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.searchClient
}

// GetMeilisearch returns the Meilisearch client
func (d *Data) GetMeilisearch() *meili.Client {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getMeilisearchUnsafe()
}

// getMeilisearchUnsafe returns Meilisearch client without locking (for internal use)
func (d *Data) getMeilisearchUnsafe() *meili.Client {
	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.MS
}

// GetElasticsearch returns the Elasticsearch client
func (d *Data) GetElasticsearch() *elastic.Client {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getElasticsearchUnsafe()
}

// getElasticsearchUnsafe returns Elasticsearch client without locking (for internal use)
func (d *Data) getElasticsearchUnsafe() *elastic.Client {
	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.ES
}

// GetOpenSearch returns the OpenSearch client
func (d *Data) GetOpenSearch() *opensearch.Client {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getOpenSearchUnsafe()
}

// getOpenSearchUnsafe returns OpenSearch client without locking (for internal use)
func (d *Data) getOpenSearchUnsafe() *opensearch.Client {
	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.OS
}

// Deprecated methods for backward compatibility

// DB returns the master database connection
// Deprecated: Use GetMasterDB() for better clarity
func (d *Data) DB() *sql.DB {
	return d.GetMasterDB()
}

// DBRead returns a slave database connection
// Deprecated: Use GetSlaveDB() for better clarity
func (d *Data) DBRead() (*sql.DB, error) {
	return d.GetSlaveDB()
}

package data

import (
	"context"
	"fmt"
	"sync"
)

// Driver interfaces define contracts for different backend types.
// Following the design pattern of database/sql, drivers register themselves
// using init() functions and are looked up at runtime based on configuration.

// DatabaseDriver defines the interface for relational database drivers.
// Implementations should handle connection lifecycle and health checks.
type DatabaseDriver interface {
	// Name returns the driver identifier (e.g., "postgres", "mysql", "sqlite")
	Name() string

	// Connect establishes a new database connection using the provided configuration.
	// The returned connection should be ready for use or return an error.
	Connect(ctx context.Context, cfg any) (any, error)

	// Close terminates the database connection and releases resources.
	Close(conn any) error

	// Ping verifies the connection is alive and functional.
	Ping(ctx context.Context, conn any) error
}

// CacheDriver defines the interface for cache/key-value store drivers.
type CacheDriver interface {
	// Name returns the driver identifier (e.g., "redis", "memcached")
	Name() string

	// Connect establishes a new cache connection.
	Connect(ctx context.Context, cfg any) (any, error)

	// Close terminates the cache connection.
	Close(conn any) error

	// Ping verifies the cache connection is alive.
	Ping(ctx context.Context, conn any) error
}

// SearchDriver defines the interface for search engine drivers.
type SearchDriver interface {
	// Name returns the driver identifier (e.g., "elasticsearch", "meilisearch")
	Name() string

	// Connect establishes a new search engine connection.
	Connect(ctx context.Context, cfg any) (any, error)

	// Close terminates the search engine connection.
	Close(conn any) error
}

// SearchEngine defines the interface that all search engine client implementations must satisfy.
// This allows the search.Client to work with any search backend through type assertions.
type SearchEngine interface {
	// Health checks if the search engine is available and responds
	Health(ctx context.Context) error

	// IndexDocument indexes a single document
	IndexDocument(ctx context.Context, index, docID string, document any) error

	// DeleteDocument deletes a document by ID
	DeleteDocument(ctx context.Context, index, docID string) error

	// IndexExists checks if an index exists
	IndexExists(ctx context.Context, index string) (bool, error)

	// CreateIndex creates a new index with optional settings
	CreateIndex(ctx context.Context, index, settings string) error
}

// MessageDriver defines the interface for message queue/broker drivers.
type MessageDriver interface {
	// Name returns the driver identifier (e.g., "kafka", "rabbitmq")
	Name() string

	// Connect establishes a new message broker connection.
	Connect(ctx context.Context, cfg any) (any, error)

	// Close terminates the message broker connection.
	Close(conn any) error
}

// StorageDriver defines the interface for object storage drivers.
type StorageDriver interface {
	// Name returns the driver identifier (e.g., "s3", "minio", "local")
	Name() string

	// Connect establishes a new storage connection.
	Connect(ctx context.Context, cfg any) (any, error)

	// Close terminates the storage connection.
	Close(conn any) error
}

// Global driver registries with mutex protection for concurrent access.
var (
	// Database drivers registry
	databaseDrivers   = make(map[string]DatabaseDriver)
	databaseDriversMu sync.RWMutex

	// Cache drivers registry
	cacheDrivers   = make(map[string]CacheDriver)
	cacheDriversMu sync.RWMutex

	// Search drivers registry
	searchDrivers   = make(map[string]SearchDriver)
	searchDriversMu sync.RWMutex

	// Message queue drivers registry
	messageDrivers   = make(map[string]MessageDriver)
	messageDriversMu sync.RWMutex

	// Storage drivers registry
	storageDrivers   = make(map[string]StorageDriver)
	storageDriversMu sync.RWMutex
)

// RegisterDatabaseDriver makes a database driver available by the provided name.
// It is intended to be called from the init function in driver packages.
//
// Example usage in a driver package:
//
//	func init() {
//	    data.RegisterDatabaseDriver(&postgresDriver{})
//	}
//
// If RegisterDatabaseDriver is called twice with the same name or if driver is nil,
// it panics.
func RegisterDatabaseDriver(driver DatabaseDriver) {
	databaseDriversMu.Lock()
	defer databaseDriversMu.Unlock()

	if driver == nil {
		panic("data: RegisterDatabaseDriver driver is nil")
	}

	name := driver.Name()
	if name == "" {
		panic("data: RegisterDatabaseDriver driver name is empty")
	}

	if _, exists := databaseDrivers[name]; exists {
		panic(fmt.Sprintf("data: RegisterDatabaseDriver called twice for driver %s", name))
	}

	databaseDrivers[name] = driver
}

// RegisterCacheDriver makes a cache driver available by the provided name.
// It follows the same pattern as RegisterDatabaseDriver.
func RegisterCacheDriver(driver CacheDriver) {
	cacheDriversMu.Lock()
	defer cacheDriversMu.Unlock()

	if driver == nil {
		panic("data: RegisterCacheDriver driver is nil")
	}

	name := driver.Name()
	if name == "" {
		panic("data: RegisterCacheDriver driver name is empty")
	}

	if _, exists := cacheDrivers[name]; exists {
		panic(fmt.Sprintf("data: RegisterCacheDriver called twice for driver %s", name))
	}

	cacheDrivers[name] = driver
}

// RegisterSearchDriver makes a search engine driver available by the provided name.
func RegisterSearchDriver(driver SearchDriver) {
	searchDriversMu.Lock()
	defer searchDriversMu.Unlock()

	if driver == nil {
		panic("data: RegisterSearchDriver driver is nil")
	}

	name := driver.Name()
	if name == "" {
		panic("data: RegisterSearchDriver driver name is empty")
	}

	if _, exists := searchDrivers[name]; exists {
		panic(fmt.Sprintf("data: RegisterSearchDriver called twice for driver %s", name))
	}

	searchDrivers[name] = driver
}

// RegisterMessageDriver makes a message queue driver available by the provided name.
func RegisterMessageDriver(driver MessageDriver) {
	messageDriversMu.Lock()
	defer messageDriversMu.Unlock()

	if driver == nil {
		panic("data: RegisterMessageDriver driver is nil")
	}

	name := driver.Name()
	if name == "" {
		panic("data: RegisterMessageDriver driver name is empty")
	}

	if _, exists := messageDrivers[name]; exists {
		panic(fmt.Sprintf("data: RegisterMessageDriver called twice for driver %s", name))
	}

	messageDrivers[name] = driver
}

// RegisterStorageDriver makes a storage driver available by the provided name.
func RegisterStorageDriver(driver StorageDriver) {
	storageDriversMu.Lock()
	defer storageDriversMu.Unlock()

	if driver == nil {
		panic("data: RegisterStorageDriver driver is nil")
	}

	name := driver.Name()
	if name == "" {
		panic("data: RegisterStorageDriver driver name is empty")
	}

	if _, exists := storageDrivers[name]; exists {
		panic(fmt.Sprintf("data: RegisterStorageDriver called twice for driver %s", name))
	}

	storageDrivers[name] = driver
}

// GetDatabaseDriver retrieves a registered database driver by name.
// It returns an error with helpful instructions if the driver is not found.
func GetDatabaseDriver(name string) (DatabaseDriver, error) {
	databaseDriversMu.RLock()
	defer databaseDriversMu.RUnlock()

	driver, ok := databaseDrivers[name]
	if !ok {
		return nil, fmt.Errorf(
			"data: database driver %q not registered\n\n"+
				"Did you forget to import the driver package?\n"+
				"Add to your imports:\n"+
				"    _ \"github.com/ncobase/ncore/data/%s\"\n\n"+
				"Available drivers: %v",
			name, name, listDatabaseDriversLocked(),
		)
	}

	return driver, nil
}

// GetCacheDriver retrieves a registered cache driver by name.
func GetCacheDriver(name string) (CacheDriver, error) {
	cacheDriversMu.RLock()
	defer cacheDriversMu.RUnlock()

	driver, ok := cacheDrivers[name]
	if !ok {
		return nil, fmt.Errorf(
			"data: cache driver %q not registered\n\n"+
				"Did you forget to import the driver package?\n"+
				"Add to your imports:\n"+
				"    _ \"github.com/ncobase/ncore/data/%s\"\n\n"+
				"Available drivers: %v",
			name, name, listCacheDriversLocked(),
		)
	}

	return driver, nil
}

// GetSearchDriver retrieves a registered search engine driver by name.
func GetSearchDriver(name string) (SearchDriver, error) {
	searchDriversMu.RLock()
	defer searchDriversMu.RUnlock()

	driver, ok := searchDrivers[name]
	if !ok {
		return nil, fmt.Errorf(
			"data: search driver %q not registered\n\n"+
				"Did you forget to import the driver package?\n"+
				"Add to your imports:\n"+
				"    _ \"github.com/ncobase/ncore/data/%s\"\n\n"+
				"Available drivers: %v",
			name, name, listSearchDriversLocked(),
		)
	}

	return driver, nil
}

// GetMessageDriver retrieves a registered message queue driver by name.
func GetMessageDriver(name string) (MessageDriver, error) {
	messageDriversMu.RLock()
	defer messageDriversMu.RUnlock()

	driver, ok := messageDrivers[name]
	if !ok {
		return nil, fmt.Errorf(
			"data: message driver %q not registered\n\n"+
				"Did you forget to import the driver package?\n"+
				"Add to your imports:\n"+
				"    _ \"github.com/ncobase/ncore/data/%s\"\n\n"+
				"Available drivers: %v",
			name, name, listMessageDriversLocked(),
		)
	}

	return driver, nil
}

// GetStorageDriver retrieves a registered storage driver by name.
func GetStorageDriver(name string) (StorageDriver, error) {
	storageDriversMu.RLock()
	defer storageDriversMu.RUnlock()

	driver, ok := storageDrivers[name]
	if !ok {
		return nil, fmt.Errorf(
			"data: storage driver %q not registered\n\n"+
				"Did you forget to import the oss module?\n"+
				"Add to your imports:\n"+
				"    \"github.com/ncobase/ncore/oss\"\n\n"+
				"Available drivers: %v",
			name, listStorageDriversLocked(),
		)
	}

	return driver, nil
}

// ListRegisteredDrivers returns a snapshot of all registered drivers.
// Useful for debugging and diagnostics.
func ListRegisteredDrivers() map[string][]string {
	result := make(map[string][]string)

	databaseDriversMu.RLock()
	result["database"] = listDatabaseDriversLocked()
	databaseDriversMu.RUnlock()

	cacheDriversMu.RLock()
	result["cache"] = listCacheDriversLocked()
	cacheDriversMu.RUnlock()

	searchDriversMu.RLock()
	result["search"] = listSearchDriversLocked()
	searchDriversMu.RUnlock()

	messageDriversMu.RLock()
	result["message"] = listMessageDriversLocked()
	messageDriversMu.RUnlock()

	storageDriversMu.RLock()
	result["storage"] = listStorageDriversLocked()
	storageDriversMu.RUnlock()

	return result
}

// Helper functions to list drivers (must be called with lock held)

func listDatabaseDriversLocked() []string {
	names := make([]string, 0, len(databaseDrivers))
	for name := range databaseDrivers {
		names = append(names, name)
	}
	return names
}

func listCacheDriversLocked() []string {
	names := make([]string, 0, len(cacheDrivers))
	for name := range cacheDrivers {
		names = append(names, name)
	}
	return names
}

func listSearchDriversLocked() []string {
	names := make([]string, 0, len(searchDrivers))
	for name := range searchDrivers {
		names = append(names, name)
	}
	return names
}

func listMessageDriversLocked() []string {
	names := make([]string, 0, len(messageDrivers))
	for name := range messageDrivers {
		names = append(names, name)
	}
	return names
}

func listStorageDriversLocked() []string {
	names := make([]string, 0, len(storageDrivers))
	for name := range storageDrivers {
		names = append(names, name)
	}
	return names
}

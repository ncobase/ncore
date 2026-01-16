package data

import (
	"context"
	"testing"
)

// Mock driver for testing
type mockDatabaseDriver struct {
	name string
}

func (d *mockDatabaseDriver) Name() string { return d.name }
func (d *mockDatabaseDriver) Connect(ctx context.Context, cfg any) (any, error) {
	return "mock-connection", nil
}
func (d *mockDatabaseDriver) Close(conn any) error                     { return nil }
func (d *mockDatabaseDriver) Ping(ctx context.Context, conn any) error { return nil }

func TestRegisterDatabaseDriver(t *testing.T) {
	// Clear registry before test
	databaseDriversMu.Lock()
	databaseDrivers = make(map[string]DatabaseDriver)
	databaseDriversMu.Unlock()

	driver := &mockDatabaseDriver{name: "test-db"}

	// Test successful registration
	RegisterDatabaseDriver(driver)

	retrieved, err := GetDatabaseDriver("test-db")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.Name() != "test-db" {
		t.Errorf("expected driver name 'test-db', got %q", retrieved.Name())
	}
}

func TestRegisterDatabaseDriverPanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when registering nil driver")
		}
	}()

	RegisterDatabaseDriver(nil)
}

func TestRegisterDatabaseDriverPanicsOnDuplicate(t *testing.T) {
	// Clear registry
	databaseDriversMu.Lock()
	databaseDrivers = make(map[string]DatabaseDriver)
	databaseDriversMu.Unlock()

	driver := &mockDatabaseDriver{name: "duplicate"}

	RegisterDatabaseDriver(driver)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when registering duplicate driver")
		}
	}()

	RegisterDatabaseDriver(driver)
}

func TestGetDatabaseDriverNotFound(t *testing.T) {
	// Clear registry
	databaseDriversMu.Lock()
	databaseDrivers = make(map[string]DatabaseDriver)
	databaseDriversMu.Unlock()

	_, err := GetDatabaseDriver("nonexistent")
	if err == nil {
		t.Errorf("expected error when getting non-existent driver")
	}

	// Error message should include helpful information
	errMsg := err.Error()
	if len(errMsg) < 50 {
		t.Errorf("error message too short, expected helpful diagnostic message")
	}
}

func TestListRegisteredDrivers(t *testing.T) {
	// Clear all registries
	databaseDriversMu.Lock()
	databaseDrivers = make(map[string]DatabaseDriver)
	databaseDriversMu.Unlock()

	cacheDriversMu.Lock()
	cacheDrivers = make(map[string]CacheDriver)
	cacheDriversMu.Unlock()

	// Register some drivers
	RegisterDatabaseDriver(&mockDatabaseDriver{name: "postgres"})
	RegisterDatabaseDriver(&mockDatabaseDriver{name: "mysql"})

	drivers := ListRegisteredDrivers()

	if len(drivers["database"]) != 2 {
		t.Errorf("expected 2 database drivers, got %d", len(drivers["database"]))
	}

	// Check that both drivers are listed
	found := make(map[string]bool)
	for _, name := range drivers["database"] {
		found[name] = true
	}

	if !found["postgres"] || !found["mysql"] {
		t.Errorf("expected to find 'postgres' and 'mysql' in registered drivers, got %v", drivers["database"])
	}
}

func TestDriverConcurrentAccess(t *testing.T) {
	// Clear registry
	databaseDriversMu.Lock()
	databaseDrivers = make(map[string]DatabaseDriver)
	databaseDriversMu.Unlock()

	RegisterDatabaseDriver(&mockDatabaseDriver{name: "concurrent-test"})

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := GetDatabaseDriver("concurrent-test")
			if err != nil {
				t.Errorf("concurrent access failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

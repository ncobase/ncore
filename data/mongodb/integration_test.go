package mongodb_test

import (
	"testing"

	"github.com/ncobase/ncore/data"
	_ "github.com/ncobase/ncore/data/mongodb" // Register the driver
)

// TestDriverRegistration verifies the MongoDB driver is properly registered
func TestDriverRegistration(t *testing.T) {
	driver, err := data.GetDatabaseDriver("mongodb")
	if err != nil {
		t.Fatalf("Failed to get MongoDB driver: %v", err)
	}

	if driver == nil {
		t.Fatal("MongoDB driver is nil")
	}

	if driver.Name() != "mongodb" {
		t.Errorf("Driver name = %v, want %v", driver.Name(), "mongodb")
	}
}

// TestListRegisteredDrivers verifies MongoDB appears in the registered drivers list
func TestListRegisteredDrivers(t *testing.T) {
	drivers := data.ListRegisteredDrivers()

	dbDrivers, ok := drivers["database"]
	if !ok {
		t.Fatal("No database drivers registered")
	}

	found := false
	for _, name := range dbDrivers {
		if name == "mongodb" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("MongoDB driver not found in registered drivers: %v", dbDrivers)
	}
}

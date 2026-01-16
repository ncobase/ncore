package meilisearch_test

import (
	"context"
	"testing"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
	_ "github.com/ncobase/ncore/data/meilisearch" // Register driver
	meilisearch "github.com/ncobase/ncore/data/meilisearch/client"
)

func TestDriverRegistration(t *testing.T) {
	// Verify driver is registered
	driver, err := data.GetSearchDriver("meilisearch")
	if err != nil {
		t.Fatalf("Failed to get meilisearch driver: %v", err)
	}

	if driver == nil {
		t.Fatal("Driver should not be nil")
	}

	if driver.Name() != "meilisearch" {
		t.Errorf("Expected driver name 'meilisearch', got '%s'", driver.Name())
	}
}

func TestDriverConnect(t *testing.T) {
	driver, err := data.GetSearchDriver("meilisearch")
	if err != nil {
		t.Fatalf("Failed to get meilisearch driver: %v", err)
	}

	// Test with empty host (should fail)
	t.Run("EmptyHost", func(t *testing.T) {
		cfg := &config.Meilisearch{
			Host:   "",
			APIKey: "",
		}

		_, err := driver.Connect(context.Background(), cfg)
		if err == nil {
			t.Error("Expected error for empty host, got nil")
		}
	})

	// Test with invalid config type (should fail)
	t.Run("InvalidConfig", func(t *testing.T) {
		_, err := driver.Connect(context.Background(), "invalid")
		if err == nil {
			t.Error("Expected error for invalid config type, got nil")
		}
	})

	// Test with valid config but unreachable host
	t.Run("UnreachableHost", func(t *testing.T) {
		cfg := &config.Meilisearch{
			Host:   "http://localhost:7700",
			APIKey: "",
		}

		conn, err := driver.Connect(context.Background(), cfg)
		// This might fail if Meilisearch is not running, which is expected
		if err != nil {
			t.Logf("Connection failed (expected if Meilisearch not running): %v", err)
			return
		}

		// If connection succeeds, verify it's the correct type
		if _, ok := conn.(*meilisearch.Client); !ok {
			t.Errorf("Expected *meili.Client, got %T", conn)
		}

		// Clean up
		if err := driver.Close(conn); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	})
}

func TestDriverClose(t *testing.T) {
	driver, err := data.GetSearchDriver("meilisearch")
	if err != nil {
		t.Fatalf("Failed to get meilisearch driver: %v", err)
	}

	// Test with invalid connection type
	t.Run("InvalidConnectionType", func(t *testing.T) {
		err := driver.Close("invalid")
		if err == nil {
			t.Error("Expected error for invalid connection type, got nil")
		}
	})

	// Test with nil client
	t.Run("ValidClient", func(t *testing.T) {
		client := meilisearch.NewMeilisearch("http://localhost:7700", "")
		err := driver.Close(client)
		if err != nil {
			t.Errorf("Close should not fail for valid client: %v", err)
		}
	})
}

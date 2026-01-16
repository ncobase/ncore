package mongodb

import (
	"context"
	"testing"

	"github.com/ncobase/ncore/data/config"
)

// TestDriverName verifies the driver returns the correct name
func TestDriverName(t *testing.T) {
	d := &driver{}
	if got := d.Name(); got != "mongodb" {
		t.Errorf("Name() = %v, want %v", got, "mongodb")
	}
}

// TestDriverConnect_InvalidConfig tests that invalid config types are rejected
func TestDriverConnect_InvalidConfig(t *testing.T) {
	d := &driver{}
	ctx := context.Background()

	// Test with nil config
	_, err := d.Connect(ctx, nil)
	if err == nil {
		t.Error("Connect() with nil config should return error")
	}

	// Test with wrong config type
	_, err = d.Connect(ctx, "invalid")
	if err == nil {
		t.Error("Connect() with invalid config type should return error")
	}
}

// TestDriverConnect_EmptyMaster tests that empty master config is rejected
func TestDriverConnect_EmptyMaster(t *testing.T) {
	d := &driver{}
	ctx := context.Background()

	// Test with nil master
	cfg := &config.MongoDB{
		Master: nil,
	}
	_, err := d.Connect(ctx, cfg)
	if err == nil {
		t.Error("Connect() with nil master should return error")
	}

	// Test with empty URI
	cfg = &config.MongoDB{
		Master: &config.MongoNode{
			URI: "",
		},
	}
	_, err = d.Connect(ctx, cfg)
	if err == nil {
		t.Error("Connect() with empty URI should return error")
	}
}

// TestDriverClose_InvalidConnection tests that invalid connection types are rejected
func TestDriverClose_InvalidConnection(t *testing.T) {
	d := &driver{}

	// Test with nil connection
	err := d.Close(nil)
	if err == nil {
		t.Error("Close() with nil connection should return error")
	}

	// Test with wrong connection type
	err = d.Close("invalid")
	if err == nil {
		t.Error("Close() with invalid connection type should return error")
	}
}

// TestDriverPing_InvalidConnection tests that invalid connection types are rejected
func TestDriverPing_InvalidConnection(t *testing.T) {
	d := &driver{}
	ctx := context.Background()

	// Test with nil connection
	err := d.Ping(ctx, nil)
	if err == nil {
		t.Error("Ping() with nil connection should return error")
	}

	// Test with wrong connection type
	err = d.Ping(ctx, "invalid")
	if err == nil {
		t.Error("Ping() with invalid connection type should return error")
	}
}

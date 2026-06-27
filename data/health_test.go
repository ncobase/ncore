package data

import (
	"context"
	"errors"
	"testing"

	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/metrics"
)

type healthRedis struct {
	err error
}

func (r healthRedis) Ping(context.Context) error {
	return r.err
}

type healthSearch struct {
	err error
}

func (s healthSearch) Health(context.Context) error {
	return s.err
}

func TestHealthUsesRedisPing(t *testing.T) {
	d := &Data{
		Conn:      &connection.Connections{RC: healthRedis{}},
		collector: metrics.NoOpCollector{},
	}

	health := d.Health(context.Background())
	redis := health["services"].(map[string]any)["redis"].(map[string]any)
	if redis["healthy"] != true {
		t.Fatalf("expected healthy redis service, got %#v", redis)
	}
}

func TestHealthReportsRedisFailure(t *testing.T) {
	d := &Data{
		Conn:      &connection.Connections{RC: healthRedis{err: errors.New("redis down")}},
		collector: metrics.NoOpCollector{},
	}

	health := d.Health(context.Background())
	if health["status"] != "degraded" {
		t.Fatalf("expected degraded health, got %#v", health)
	}
	redis := health["services"].(map[string]any)["redis"].(map[string]any)
	if redis["healthy"] != false || redis["error"] == "" {
		t.Fatalf("expected redis error details, got %#v", redis)
	}
}

func TestHealthReportsSearchEngines(t *testing.T) {
	d := &Data{
		Conn: &connection.Connections{
			MS: healthSearch{},
			ES: healthSearch{err: errors.New("es down")},
		},
		collector: metrics.NoOpCollector{},
	}

	health := d.Health(context.Background())
	if health["status"] != "degraded" {
		t.Fatalf("expected degraded health, got %#v", health)
	}

	search := health["services"].(map[string]any)["search"].(map[string]any)
	if search["healthy"] != false {
		t.Fatalf("expected unhealthy search summary, got %#v", search)
	}
	engines := search["engines"].(map[string]any)
	if engines["meilisearch"].(map[string]any)["healthy"] != true {
		t.Fatalf("expected meilisearch healthy, got %#v", engines["meilisearch"])
	}
	if engines["elasticsearch"].(map[string]any)["healthy"] != false {
		t.Fatalf("expected elasticsearch unhealthy, got %#v", engines["elasticsearch"])
	}
}

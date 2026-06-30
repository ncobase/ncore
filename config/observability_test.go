package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestTracingConfigUsesTopLevelKeys(t *testing.T) {
	v := viper.New()
	v.Set("tracing.endpoint", "otel-collector:4317")
	v.Set("tracing.service_name", "api")
	v.Set("tracing.sampling_rate", 0.25)
	v.Set("tracing.batch_timeout", "2s")
	v.Set("tracing.headers.authorization", "Bearer token")

	cfg := GetTracingConfig(v)
	if cfg == nil {
		t.Fatal("expected tracing config")
	}
	if cfg.Endpoint != "otel-collector:4317" {
		t.Fatalf("unexpected endpoint: %s", cfg.Endpoint)
	}
	if cfg.ServiceName != "api" {
		t.Fatalf("unexpected service name: %s", cfg.ServiceName)
	}
	if cfg.SamplingRate != 0.25 {
		t.Fatalf("unexpected sampling rate: %f", cfg.SamplingRate)
	}
	if cfg.BatchTimeout != 2*time.Second {
		t.Fatalf("unexpected batch timeout: %s", cfg.BatchTimeout)
	}
	if cfg.MaxExportBatchSize != 512 {
		t.Fatalf("unexpected max export batch size: %d", cfg.MaxExportBatchSize)
	}
	if cfg.Headers["authorization"] != "Bearer token" {
		t.Fatalf("unexpected headers: %#v", cfg.Headers)
	}
}

func TestTracingConfigPreservesZeroSamplingRate(t *testing.T) {
	v := viper.New()
	v.Set("tracing.endpoint", "otel-collector:4317")
	v.Set("tracing.sampling_rate", 0)

	cfg := GetTracingConfig(v)
	if cfg == nil {
		t.Fatal("expected tracing config")
	}
	if cfg.SamplingRate != 0 {
		t.Fatalf("unexpected sampling rate: %f", cfg.SamplingRate)
	}
}

func TestObservabilityConfigDoesNotReadLegacyObservesKeys(t *testing.T) {
	v := viper.New()
	v.Set("observes.tracer.endpoint", "legacy:4317")
	v.Set("observes.sentry.endpoint", "https://legacy.example/sentry")

	if cfg := GetTracingConfig(v); cfg != nil {
		t.Fatalf("legacy observes.tracer produced tracing config: %#v", cfg)
	}
	if cfg := GetSentryConfig(v); cfg != nil {
		t.Fatalf("legacy observes.sentry produced sentry config: %#v", cfg)
	}
}

func TestSentryConfigUsesTopLevelKeys(t *testing.T) {
	v := viper.New()
	v.Set("sentry.endpoint", "https://example@sentry.io/1")
	v.Set("sentry.environment", "production")
	v.Set("sentry.release", "v1.2.3")
	v.Set("sentry.sample_rate", 0.5)

	cfg := GetSentryConfig(v)
	if cfg == nil {
		t.Fatal("expected sentry config")
	}
	if cfg.Endpoint != "https://example@sentry.io/1" {
		t.Fatalf("unexpected endpoint: %s", cfg.Endpoint)
	}
	if cfg.Environment != "production" {
		t.Fatalf("unexpected environment: %s", cfg.Environment)
	}
	if cfg.Release != "v1.2.3" {
		t.Fatalf("unexpected release: %s", cfg.Release)
	}
	if cfg.SampleRate != 0.5 {
		t.Fatalf("unexpected sample rate: %f", cfg.SampleRate)
	}
}

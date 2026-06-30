package tracing

import (
	"testing"
	"time"
)

func TestNormalizeTracerOptionDefaults(t *testing.T) {
	opt := &TracerOption{}

	normalizeTracerOption(opt)

	if opt.Name != "application" {
		t.Fatalf("unexpected default name: %s", opt.Name)
	}
	if opt.SamplingRate != 0 {
		t.Fatalf("unexpected default sampling rate: %f", opt.SamplingRate)
	}
	if opt.MaxAttributes != 128 {
		t.Fatalf("unexpected default max attributes: %d", opt.MaxAttributes)
	}
	if opt.MaxAttributeLength != 1024 {
		t.Fatalf("unexpected default max attribute length: %d", opt.MaxAttributeLength)
	}
	if opt.MaxEventsPerSpan != 128 {
		t.Fatalf("unexpected default max events per span: %d", opt.MaxEventsPerSpan)
	}
	if opt.MaxLinksPerSpan != 128 {
		t.Fatalf("unexpected default max links per span: %d", opt.MaxLinksPerSpan)
	}
	if opt.MaxQueueSize != 2048 {
		t.Fatalf("unexpected default max queue size: %d", opt.MaxQueueSize)
	}
	if opt.BatchTimeout != 5*time.Second {
		t.Fatalf("unexpected default batch timeout: %s", opt.BatchTimeout)
	}
	if opt.ExportTimeout != 30*time.Second {
		t.Fatalf("unexpected default export timeout: %s", opt.ExportTimeout)
	}
	if opt.MaxExportBatchSize != 512 {
		t.Fatalf("unexpected default max export batch size: %d", opt.MaxExportBatchSize)
	}
}

func TestNormalizeTracerOptionPreservesExplicitValues(t *testing.T) {
	opt := &TracerOption{
		Name:               "api",
		SamplingRate:       0.25,
		MaxAttributes:      64,
		MaxAttributeLength: 256,
		MaxEventsPerSpan:   32,
		MaxLinksPerSpan:    16,
		MaxQueueSize:       1024,
		BatchTimeout:       time.Second,
		ExportTimeout:      2 * time.Second,
		MaxExportBatchSize: 128,
	}

	normalizeTracerOption(opt)

	if opt.Name != "api" || opt.SamplingRate != 0.25 || opt.MaxAttributes != 64 ||
		opt.MaxAttributeLength != 256 || opt.MaxEventsPerSpan != 32 ||
		opt.MaxLinksPerSpan != 16 || opt.MaxQueueSize != 1024 ||
		opt.BatchTimeout != time.Second || opt.ExportTimeout != 2*time.Second ||
		opt.MaxExportBatchSize != 128 {
		t.Fatalf("explicit tracing options were not preserved: %#v", opt)
	}
}

func TestExporterOptions(t *testing.T) {
	opt := &TracerOption{
		URL:           "http://collector.example:4317",
		ExportTimeout: time.Second,
		Headers: map[string]string{
			"authorization": "Bearer token",
		},
	}
	normalizeTracerOption(opt)

	options, err := buildExporterOptions(opt)
	if err != nil {
		t.Fatalf("buildExporterOptions returned unexpected error: %v", err)
	}
	if len(options) != 4 {
		t.Fatalf("unexpected exporter option count: %d", len(options))
	}

	opt.TLSEnabled = true
	options, err = buildExporterOptions(opt)
	if err != nil {
		t.Fatalf("buildExporterOptions with TLS returned unexpected error: %v", err)
	}
	if len(options) != 4 {
		t.Fatalf("unexpected TLS exporter option count: %d", len(options))
	}
}

func TestExporterOptionsUseSchemeSecurity(t *testing.T) {
	opt := &TracerOption{
		URL:           "https://collector.example:4317",
		ExportTimeout: time.Second,
	}
	normalizeTracerOption(opt)

	options, err := buildExporterOptions(opt)
	if err != nil {
		t.Fatalf("buildExporterOptions returned unexpected error: %v", err)
	}
	if len(options) != 2 {
		t.Fatalf("unexpected HTTPS exporter option count: %d", len(options))
	}
}

func TestTransportCredentialsRequiresCertAndKeyTogether(t *testing.T) {
	_, err := transportCredentials(&TracerOption{TLSEnabled: true, TLSCertFile: "client.crt"})
	if err == nil {
		t.Fatal("expected missing TLS key error")
	}
}

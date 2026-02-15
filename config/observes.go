package config

import (
	"time"

	"github.com/spf13/viper"
)

// Sentry config struct
type Sentry struct {
	Endpoint    string  `json:"endpoint" yaml:"endpoint"`
	Environment string  `json:"environment" yaml:"environment"`
	Release     string  `json:"release" yaml:"release"`
	SampleRate  float64 `json:"sample_rate" yaml:"sample_rate"`
}

// getSentryConfig get sentry config
func getSentryConfig(v *viper.Viper) *Sentry {
	return &Sentry{
		Endpoint:    v.GetString("observes.sentry.endpoint"),
		Environment: v.GetString("observes.sentry.environment"),
		Release:     v.GetString("observes.sentry.release"),
		SampleRate:  getFloat64OrDefault(v, "observes.sentry.sample_rate", 1.0),
	}
}

// Tracer config struct for OpenTelemetry
type Tracer struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"` // OTLP gRPC endpoint

	// Service identification
	ServiceName    string `json:"service_name" yaml:"service_name"`
	ServiceVersion string `json:"service_version" yaml:"service_version"`
	Environment    string `json:"environment" yaml:"environment"`

	// Sampling configuration
	SamplingRate float64 `json:"sampling_rate" yaml:"sampling_rate"` // 0.0 to 1.0

	// Performance tuning
	MaxExportBatchSize int           `json:"max_export_batch_size" yaml:"max_export_batch_size"`
	BatchTimeout       time.Duration `json:"batch_timeout" yaml:"batch_timeout"`
	ExportTimeout      time.Duration `json:"export_timeout" yaml:"export_timeout"`
	MaxQueueSize       int           `json:"max_queue_size" yaml:"max_queue_size"`

	// Attribute limits
	MaxAttributes      int `json:"max_attributes" yaml:"max_attributes"`
	MaxAttributeLength int `json:"max_attribute_length" yaml:"max_attribute_length"`
	MaxEventsPerSpan   int `json:"max_events_per_span" yaml:"max_events_per_span"`
	MaxLinksPerSpan    int `json:"max_links_per_span" yaml:"max_links_per_span"`

	// TLS configuration
	TLSEnabled         bool   `json:"tls_enabled" yaml:"tls_enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	TLSCertFile        string `json:"tls_cert_file" yaml:"tls_cert_file"`
	TLSKeyFile         string `json:"tls_key_file" yaml:"tls_key_file"`
	TLSCAFile          string `json:"tls_ca_file" yaml:"tls_ca_file"`

	// Headers for authentication
	Headers map[string]string `json:"headers" yaml:"headers"`
}

// getTracerConfig get tracer config with defaults
func getTracerConfig(v *viper.Viper) *Tracer {
	return &Tracer{
		Endpoint: v.GetString("observes.tracer.endpoint"),

		// Service identification
		ServiceName:    v.GetString("observes.tracer.service_name"),
		ServiceVersion: v.GetString("observes.tracer.service_version"),
		Environment:    v.GetString("observes.tracer.environment"),

		// Sampling with default to 100% in development
		SamplingRate: getFloat64OrDefault(v, "observes.tracer.sampling_rate", 1.0),

		// Performance tuning with sensible defaults
		MaxExportBatchSize: getIntOrDefault(v, "observes.tracer.max_export_batch_size", 512),
		BatchTimeout:       getDurationOrDefault(v, "observes.tracer.batch_timeout", 5*time.Second),
		ExportTimeout:      getDurationOrDefault(v, "observes.tracer.export_timeout", 30*time.Second),
		MaxQueueSize:       getIntOrDefault(v, "observes.tracer.max_queue_size", 2048),

		// Attribute limits with defaults
		MaxAttributes:      getIntOrDefault(v, "observes.tracer.max_attributes", 128),
		MaxAttributeLength: getIntOrDefault(v, "observes.tracer.max_attribute_length", 1024),
		MaxEventsPerSpan:   getIntOrDefault(v, "observes.tracer.max_events_per_span", 128),
		MaxLinksPerSpan:    getIntOrDefault(v, "observes.tracer.max_links_per_span", 128),

		// TLS configuration
		TLSEnabled:         v.GetBool("observes.tracer.tls_enabled"),
		InsecureSkipVerify: v.GetBool("observes.tracer.insecure_skip_verify"),
		TLSCertFile:        v.GetString("observes.tracer.tls_cert_file"),
		TLSKeyFile:         v.GetString("observes.tracer.tls_key_file"),
		TLSCAFile:          v.GetString("observes.tracer.tls_ca_file"),

		// Headers for authentication
		Headers: v.GetStringMapString("observes.tracer.headers"),
	}
}

// Observes config struct
type Observes struct {
	Sentry *Sentry
	Tracer *Tracer
}

// get Observes config
func getObservesConfig(v *viper.Viper) *Observes {
	return &Observes{
		Sentry: getSentryConfig(v),
		Tracer: getTracerConfig(v),
	}
}

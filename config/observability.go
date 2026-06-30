package config

import (
	"time"

	"github.com/spf13/viper"
)

// Sentry contains Sentry error tracking infrastructure settings.
type Sentry struct {
	Endpoint    string  `json:"endpoint" yaml:"endpoint"`
	Environment string  `json:"environment" yaml:"environment"`
	Release     string  `json:"release" yaml:"release"`
	SampleRate  float64 `json:"sample_rate" yaml:"sample_rate"`
}

// Tracing contains OpenTelemetry tracing infrastructure settings.
type Tracing struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	ServiceName    string `json:"service_name" yaml:"service_name"`
	ServiceVersion string `json:"service_version" yaml:"service_version"`
	Environment    string `json:"environment" yaml:"environment"`

	SamplingRate float64 `json:"sampling_rate" yaml:"sampling_rate"`

	MaxExportBatchSize int           `json:"max_export_batch_size" yaml:"max_export_batch_size"`
	BatchTimeout       time.Duration `json:"batch_timeout" yaml:"batch_timeout"`
	ExportTimeout      time.Duration `json:"export_timeout" yaml:"export_timeout"`
	MaxQueueSize       int           `json:"max_queue_size" yaml:"max_queue_size"`

	MaxAttributes      int `json:"max_attributes" yaml:"max_attributes"`
	MaxAttributeLength int `json:"max_attribute_length" yaml:"max_attribute_length"`
	MaxEventsPerSpan   int `json:"max_events_per_span" yaml:"max_events_per_span"`
	MaxLinksPerSpan    int `json:"max_links_per_span" yaml:"max_links_per_span"`

	TLSEnabled         bool   `json:"tls_enabled" yaml:"tls_enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	TLSCertFile        string `json:"tls_cert_file" yaml:"tls_cert_file"`
	TLSKeyFile         string `json:"tls_key_file" yaml:"tls_key_file"`
	TLSCAFile          string `json:"tls_ca_file" yaml:"tls_ca_file"`

	Headers map[string]string `json:"headers" yaml:"headers"`
}

// GetSentryConfig returns Sentry configuration from viper when configured.
func GetSentryConfig(v *viper.Viper) *Sentry {
	if !v.IsSet("sentry") {
		return nil
	}
	return &Sentry{
		Endpoint:    v.GetString("sentry.endpoint"),
		Environment: v.GetString("sentry.environment"),
		Release:     v.GetString("sentry.release"),
		SampleRate:  getFloat64OrDefault(v, "sentry.sample_rate", 1.0),
	}
}

// GetTracingConfig returns OpenTelemetry tracing configuration from viper when configured.
func GetTracingConfig(v *viper.Viper) *Tracing {
	if !v.IsSet("tracing") {
		return nil
	}
	return &Tracing{
		Endpoint: v.GetString("tracing.endpoint"),

		ServiceName:    v.GetString("tracing.service_name"),
		ServiceVersion: v.GetString("tracing.service_version"),
		Environment:    v.GetString("tracing.environment"),

		SamplingRate: getFloat64OrDefault(v, "tracing.sampling_rate", 1.0),

		MaxExportBatchSize: getIntOrDefault(v, "tracing.max_export_batch_size", 512),
		BatchTimeout:       getDurationOrDefault(v, "tracing.batch_timeout", 5*time.Second),
		ExportTimeout:      getDurationOrDefault(v, "tracing.export_timeout", 30*time.Second),
		MaxQueueSize:       getIntOrDefault(v, "tracing.max_queue_size", 2048),

		MaxAttributes:      getIntOrDefault(v, "tracing.max_attributes", 128),
		MaxAttributeLength: getIntOrDefault(v, "tracing.max_attribute_length", 1024),
		MaxEventsPerSpan:   getIntOrDefault(v, "tracing.max_events_per_span", 128),
		MaxLinksPerSpan:    getIntOrDefault(v, "tracing.max_links_per_span", 128),

		TLSEnabled:         v.GetBool("tracing.tls_enabled"),
		InsecureSkipVerify: v.GetBool("tracing.insecure_skip_verify"),
		TLSCertFile:        v.GetString("tracing.tls_cert_file"),
		TLSKeyFile:         v.GetString("tracing.tls_key_file"),
		TLSCAFile:          v.GetString("tracing.tls_ca_file"),

		Headers: v.GetStringMapString("tracing.headers"),
	}
}

func getSentryConfig(v *viper.Viper) *Sentry {
	return GetSentryConfig(v)
}

func getTracingConfig(v *viper.Viper) *Tracing {
	return GetTracingConfig(v)
}

package config

import "github.com/google/wire"

// ProviderSet is the wire provider set for the config package.
// It provides the main *Config and extracts sub-configurations for
// other modules to use.
//
// Usage:
//
//	wire.Build(
//	    config.ProviderSet,
//	    // ... other providers
//	)
//
// Available configurations:
//   - *Config: Main configuration
//   - *Logger: Logger configuration
//   - *Data: Data layer configuration
//   - *Extension: Extension system configuration
//   - *Auth: Authentication configuration
//   - *Storage: Storage configuration
//   - *Email: Email configuration
//   - *OAuth: OAuth configuration
//   - *Tracing: OpenTelemetry tracing configuration
//   - *Sentry: Sentry error tracking configuration
var ProviderSet = wire.NewSet(
	GetConfig,
	ProvideLoggerConfig,
	ProvideDataConfig,
	ProvideExtensionConfig,
	ProvideAuthConfig,
	ProvideStorageConfig,
	ProvideEmailConfig,
	ProvideOAuthConfig,
	ProvideTracingConfig,
	ProvideSentryConfig,
)

// ProvideLoggerConfig provides the logger configuration.
func ProvideLoggerConfig(cfg *Config) *Logger {
	if cfg == nil {
		return nil
	}
	return cfg.Logger
}

// ProvideDataConfig provides the data layer configuration.
func ProvideDataConfig(cfg *Config) *Data {
	if cfg == nil {
		return nil
	}
	return cfg.Data
}

// ProvideExtensionConfig provides the extension system configuration.
func ProvideExtensionConfig(cfg *Config) *Extension {
	if cfg == nil {
		return nil
	}
	return cfg.Extension
}

// ProvideAuthConfig provides the authentication configuration.
func ProvideAuthConfig(cfg *Config) *Auth {
	if cfg == nil {
		return nil
	}
	return cfg.Auth
}

// ProvideStorageConfig provides the storage configuration.
func ProvideStorageConfig(cfg *Config) *Storage {
	if cfg == nil {
		return nil
	}
	return cfg.Storage
}

// ProvideEmailConfig provides the email configuration.
func ProvideEmailConfig(cfg *Config) *Email {
	if cfg == nil {
		return nil
	}
	return cfg.Email
}

// ProvideOAuthConfig provides the OAuth configuration.
func ProvideOAuthConfig(cfg *Config) *OAuth {
	if cfg == nil {
		return nil
	}
	return cfg.OAuth
}

// ProvideTracingConfig provides OpenTelemetry tracing configuration.
func ProvideTracingConfig(cfg *Config) *Tracing {
	if cfg == nil {
		return nil
	}
	return cfg.Tracing
}

// ProvideSentryConfig provides Sentry error tracking configuration.
func ProvideSentryConfig(cfg *Config) *Sentry {
	if cfg == nil {
		return nil
	}
	return cfg.Sentry
}

// Package config provides centralized configuration management for ncore applications
// using Viper with support for multiple formats, environment variables, and hot-reloading.
//
// This package manages application configuration including:
//   - Server settings (HTTP, gRPC)
//   - Database connections (PostgreSQL, MySQL, MongoDB, Redis)
//   - Message queues (Kafka, RabbitMQ)
//   - Search engines (Elasticsearch, OpenSearch, Meilisearch)
//   - Storage (Object storage, filesystem)
//   - Security (JWT, encryption)
//   - Observability (logging, tracing, metrics, Sentry)
//   - Email services (SMTP, SendGrid, Mailgun, etc.)
//
// # Configuration Loading
//
// Load configuration from file:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Load with custom path:
//
//	cfg, err := config.LoadWithPath("./config/app.yaml")
//
// # Configuration Format
//
// Supports YAML, JSON, and TOML formats. Example YAML:
//
//	server:
//	  host: 0.0.0.0
//	  port: 8080
//	  mode: production
//
//	database:
//	  master:
//	    driver: postgres
//	    host: localhost
//	    port: 5432
//	    database: myapp
//
//	jwt:
//	  secret: your-secret-key
//	  expire: 24h
//
// # Environment Variables
//
// Override config values with environment variables using underscores:
//
//	export SERVER_PORT=9000
//	export DATABASE_MASTER_HOST=db.example.com
//	export JWT_SECRET=production-secret
//
// Environment variables take precedence over file configuration.
//
// # Hot Reloading
//
// Watch configuration file for changes:
//
//	config.WatchConfig(func(cfg *config.Config) {
//	    log.Println("Configuration reloaded")
//	    // React to configuration changes
//	})
//
// # Default Values
//
// The package provides sensible defaults for all settings:
//   - Server: port 8080, debug mode
//   - Database: localhost connections
//   - JWT: 24h expiration
//   - Connection pools: optimized sizes
//   - Timeouts: reasonable defaults
//
// All helper functions (getDurationOrDefault, getIntOrDefault, etc.)
// automatically fall back to defaults when values are not specified.
//
// # Provider Sets
//
// The package exports Wire provider sets for dependency injection:
//
//	ProviderSet // Provides complete Config
//
// Use with Wire for automatic configuration wiring in applications.
package config

// Package rabbitmq provides a RabbitMQ driver for ncore/data.
//
// This driver uses amqp091-go (github.com/rabbitmq/amqp091-go) as the underlying
// AMQP 0-9-1 client library. It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/rabbitmq"
//
// The driver supports RabbitMQ connection configuration including connection URLs,
// virtual hosts, and connection parameters.
package rabbitmq

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// driver implements data.MessageDriver for RabbitMQ.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "rabbitmq"
}

// Connect establishes a RabbitMQ connection using the provided configuration.
//
// The configuration must be a *config.RabbitMQ containing:
//   - URL: AMQP connection URL (e.g., "amqp://user:pass@localhost:5672/") or Host:Port
//   - VHost: Virtual host name (default: "/")
//
// Example URLs:
//
//	"amqp://guest:guest@localhost:5672/"        // Default guest credentials
//	"amqp://user:pass@rabbitmq.example.com/"    // Custom credentials
//	"amqps://user:pass@rabbitmq.example.com/"   // TLS enabled
//
// Returns an *amqp.Connection that can be used to create channels for
// publishing and consuming messages.
func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	rmqCfg, ok := cfg.(*config.RabbitMQ)
	if !ok {
		return nil, fmt.Errorf("rabbitmq: invalid configuration type, expected *config.RabbitMQ")
	}

	connURL := rmqCfg.URL
	if connURL == "" {
		return nil, fmt.Errorf("rabbitmq: URL is empty")
	}

	// Check if URL has scheme, if not construct it
	if !strings.HasPrefix(connURL, "amqp://") && !strings.HasPrefix(connURL, "amqps://") {
		// Assume URL field contains host:port
		u := url.URL{
			Scheme: "amqp",
			Host:   connURL,
		}

		if rmqCfg.Username != "" || rmqCfg.Password != "" {
			u.User = url.UserPassword(rmqCfg.Username, rmqCfg.Password)
		}

		if rmqCfg.Vhost != "" {
			if !strings.HasPrefix(rmqCfg.Vhost, "/") {
				u.Path = "/" + rmqCfg.Vhost
			} else {
				u.Path = "/" + rmqCfg.Vhost
			}
		}

		connURL = u.String()
	}

	conn, err := amqp.Dial(connURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: failed to connect: %w", err)
	}

	return NewRabbitMQ(conn), nil
}

// Close terminates the RabbitMQ connection and releases resources.
func (d *driver) Close(conn any) error {
	amqpConn, ok := conn.(*amqp.Connection)
	if !ok {
		return fmt.Errorf("rabbitmq: invalid connection type, expected *amqp.Connection")
	}

	if err := amqpConn.Close(); err != nil {
		return fmt.Errorf("rabbitmq: failed to close connection: %w", err)
	}

	return nil
}

// init registers the RabbitMQ driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterMessageDriver(&driver{})
}

// Package service contains business logic for the MongoDB example API.
package service

import (
	"github.com/ncobase/ncore/examples/02-mongodb-api/data"
	"github.com/ncobase/ncore/logging/logger"
)

// Service aggregates all business logic services.
type Service struct {
	User *UserService
}

// NewService creates a new service instance with all sub-services initialized.
func NewService(d *data.Data, logger *logger.Logger) *Service {
	return &Service{
		User: NewUserService(d, logger),
	}
}

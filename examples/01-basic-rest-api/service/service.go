// Package service contains business logic for the basic REST API example.
package service

import (
	"github.com/ncobase/ncore/examples/01-basic-rest-api/data"
	"github.com/ncobase/ncore/logging/logger"
)

// Service aggregates all business logic services.
type Service struct {
	Task *TaskService
}

// NewService creates a new service instance with all sub-services initialized.
func NewService(d *data.Data, logger *logger.Logger) *Service {
	return &Service{
		Task: NewTaskService(d, logger),
	}
}

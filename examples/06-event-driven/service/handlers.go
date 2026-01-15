package service

import (
	"context"

	"github.com/ncobase/ncore/examples/06-event-driven/event"
	"github.com/ncobase/ncore/logging/logger"
)

// NotificationService handles sending notifications based on events.
type NotificationService struct {
	logger *logger.Logger
}

// NewNotificationService creates a new notification service.
func NewNotificationService(logger *logger.Logger) *NotificationService {
	return &NotificationService{
		logger: logger,
	}
}

// HandleUserRegistered handles user registered events.
func (s *NotificationService) HandleUserRegistered(ctx context.Context, evt *event.Event) error {
	email, _ := evt.Payload["email"].(string)
	name, _ := evt.Payload["name"].(string)

	s.logger.Info(ctx, "Sending welcome email",
		"to", email,
		"name", name,
		"event_id", evt.ID)

	// Simulate email sending
	// In real implementation, this would send actual email

	return nil
}

// HandleUserUpdated handles user updated events.
func (s *NotificationService) HandleUserUpdated(ctx context.Context, evt *event.Event) error {
	email, _ := evt.Payload["new_email"].(string)

	s.logger.Info(ctx, "Sending profile update notification",
		"to", email,
		"event_id", evt.ID)

	return nil
}

// AnalyticsService handles analytics based on events.
type AnalyticsService struct {
	logger *logger.Logger
	events []string // Store event IDs for demo
}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService(logger *logger.Logger) *AnalyticsService {
	return &AnalyticsService{
		logger: logger,
		events: make([]string, 0),
	}
}

// HandleEvent handles any event for analytics.
func (s *AnalyticsService) HandleEvent(ctx context.Context, evt *event.Event) error {
	s.events = append(s.events, evt.ID)

	s.logger.Info(ctx, "Tracking event",
		"type", evt.Type,
		"aggregate_id", evt.AggregateID,
		"event_id", evt.ID)

	return nil
}

// GetStats returns analytics statistics.
func (s *AnalyticsService) GetStats() map[string]any {
	return map[string]any{
		"total_events_tracked": len(s.events),
	}
}

// AuditService handles audit logging based on events.
type AuditService struct {
	logger *logger.Logger
}

// NewAuditService creates a new audit service.
func NewAuditService(logger *logger.Logger) *AuditService {
	return &AuditService{
		logger: logger,
	}
}

// HandleEvent handles any event for audit logging.
func (s *AuditService) HandleEvent(ctx context.Context, evt *event.Event) error {
	s.logger.Info(ctx, "Audit log",
		"event_type", evt.Type,
		"event_id", evt.ID,
		"aggregate_id", evt.AggregateID,
		"timestamp", evt.Timestamp,
	)

	return nil
}

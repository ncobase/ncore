// Package service handles notification delivery for the full app.
package service

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/full-application/internal/event"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/messaging/email"
)

type Service struct {
	sender email.Sender
	logger *logger.Logger
}

func NewService(logger *logger.Logger, sender email.Sender) *Service {
	return &Service{
		sender: sender,
		logger: logger,
	}
}

func (s *Service) HandleTaskCreated(ctx context.Context, evt *event.Event) error {
	workspaceID := evt.WorkspaceID
	taskID := evt.Payload["task_id"].(string)
	title := evt.Payload["title"].(string)

	subject := "New Task Created"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>New Task Created</h2>
			<p>A new task has been created in your workspace:</p>
			<ul>
				<li><strong>Title:</strong> %s</li>
				<li><strong>Task ID:</strong> %s</li>
				<li><strong>Workspace ID:</strong> %s</li>
			</ul>
			<p>Created by: %s</p>
		</body>
		</html>
	`, title, taskID, workspaceID, evt.UserID)

	return s.sendNotification(ctx, subject, body, []string{})
}

func (s *Service) HandleTaskAssigned(ctx context.Context, evt *event.Event) error {
	taskID := evt.Payload["task_id"].(string)
	title := evt.Payload["title"].(string)
	assigneeID := evt.Payload["assigned_to"].(string)

	subject := "Task Assigned to You"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Task Assigned</h2>
			<p>A task has been assigned to you:</p>
			<ul>
				<li><strong>Title:</strong> %s</li>
				<li><strong>Task ID:</strong> %s</li>
			</ul>
		</body>
		</html>
	`, title, taskID)

	return s.sendNotification(ctx, subject, body, []string{assigneeID})
}

func (s *Service) HandleCommentCreated(ctx context.Context, evt *event.Event) error {
	taskID := evt.Payload["task_id"].(string)
	content := evt.Payload["content"].(string)

	subject := "New Comment on Task"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>New Comment</h2>
			<p>A new comment has been added:</p>
			<ul>
				<li><strong>Task ID:</strong> %s</li>
			</ul>
			<p><strong>Comment:</strong><br/>%s</p>
			<p>Commented by: %s</p>
		</body>
		</html>
	`, taskID, content, evt.UserID)

	return s.sendNotification(ctx, subject, body, []string{})
}

func (s *Service) sendNotification(ctx context.Context, subject, body string, to []string) error {
	if s.sender == nil {
		s.logger.Warn(ctx, "Email sender not configured, skipping notification")
		return nil
	}
	if len(to) == 0 {
		s.logger.Info(ctx, "Notification skipped: no recipients", "subject", subject)
		return nil
	}

	template := email.Template{
		Subject:  subject,
		Template: body,
		Keyword:  subject,
		URL:      "",
	}

	for _, recipient := range to {
		if _, err := s.sender.SendTemplateEmail(recipient, template); err != nil {
			s.logger.Error(ctx, "Failed to send notification", "error", err, "recipient", recipient)
		}
	}

	return nil
}

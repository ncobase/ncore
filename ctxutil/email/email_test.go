package email

import (
	"context"
	"errors"
	"testing"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/ctxutil"
	mail "github.com/ncobase/ncore/messaging/email"
)

type mockSender struct {
	messageID string
	err       error
}

func (m *mockSender) SendTemplateEmail(_ string, _ mail.Template) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.messageID, nil
}

func TestEmailHelpers(t *testing.T) {
	sender := &mockSender{messageID: "message-1"}
	ctx := SetSender(context.Background(), sender)

	gotSender, err := GetSender(ctx)
	if err != nil {
		t.Fatalf("GetSender returned unexpected error: %v", err)
	}
	if gotSender != sender {
		t.Fatalf("GetSender returned unexpected sender instance: %#v", gotSender)
	}

	messageID, err := SendWithTemplate(ctx, "a@example.com", mail.Template{Subject: "s"})
	if err != nil {
		t.Fatalf("SendWithTemplate returned unexpected error: %v", err)
	}
	if messageID != "message-1" {
		t.Fatalf("unexpected message ID: %q", messageID)
	}

	ctxUnknown := ctxutil.SetConfig(context.Background(), &config.Config{
		Email: &config.Email{Provider: "unknown-provider"},
	})
	if _, err := GetSender(ctxUnknown); err == nil {
		t.Fatal("expected unknown provider error")
	}

	ctxMissingConfig := ctxutil.SetConfig(context.Background(), nil)
	if _, err := GetSender(ctxMissingConfig); err == nil {
		t.Fatal("expected missing config error")
	}

	ctxSendErr := SetSender(context.Background(), &mockSender{err: errors.New("send failure")})
	if _, err := SendWithTemplate(ctxSendErr, "a@example.com", mail.Template{Subject: "s"}); err == nil {
		t.Fatal("expected send error")
	}
}

func TestToProviderConfig(t *testing.T) {
	cfg, err := ToProviderConfig(&config.Email{
		Provider: "smtp",
		SMTP: &config.SMTPConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: "587",
			Username: "user",
			Password: "secret",
			From:     "noreply@example.com",
		},
	})
	if err != nil {
		t.Fatalf("ToProviderConfig returned unexpected error: %v", err)
	}
	smtp, ok := cfg.(*mail.SMTPConfig)
	if !ok {
		t.Fatalf("expected SMTP config, got %#v", cfg)
	}
	if smtp.SMTPHost != "smtp.example.com" || smtp.From != "noreply@example.com" {
		t.Fatalf("unexpected SMTP config: %#v", smtp)
	}
}

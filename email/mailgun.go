package email

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

// MailgunConfig holds the configuration for Mailgun
type MailgunConfig struct {
	Key    string
	Domain string
	From   string
}

// MailgunSender implements EmailSender for Mailgun
type MailgunSender struct {
	Config *MailgunConfig
}

func (s *MailgunSender) SendTemplateEmail(recipientEmail string, template AuthEmailTemplate) (string, error) {
	mg := mailgun.NewMailgun(s.Config.Domain, s.Config.Key)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message := mg.NewMessage(s.Config.From, template.Subject, "")
	message.SetTemplate(template.Template)
	_ = message.AddRecipient(recipientEmail)
	message.AddVariable("keyword", template.Keyword)
	message.AddVariable("url", template.URL)

	_, id, err := mg.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return "", err
	}

	log.Printf("Email queued: %s", id)
	return id, nil
}

func validateMailgunConfig(config *MailgunConfig) error {
	if config.Key == "" || config.Domain == "" || config.From == "" {
		return errors.New("invalid Mailgun configuration")
	}
	return nil
}

package email

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridConfig holds the configuration for SendGrid
type SendGridConfig struct {
	Key  string
	From string
}

// SendGridSender implements EmailSender for SendGrid
type SendGridSender struct {
	Config *SendGridConfig
}

func (s *SendGridSender) SendTemplateEmail(recipientEmail string, template EmailTemplate) (string, error) {
	from := mail.NewEmail("Example User", s.Config.From)
	subject := template.Subject
	to := mail.NewEmail("Recipient", recipientEmail)
	plainTextContent := fmt.Sprintf("Keyword: %s\nURL: %s", template.Keyword, template.URL)
	htmlContent := fmt.Sprintf("<strong>Keyword:</strong> %s<br><strong>URL:</strong> %s", template.Keyword, template.URL)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	client := sendgrid.NewSendClient(s.Config.Key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.SendWithContext(ctx, message)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return "", err
	}

	if response.StatusCode != 202 {
		err := fmt.Errorf("failed to send email, status code: %d", response.StatusCode)
		log.Printf("Error sending email: %v", err)
		return "", err
	}

	log.Printf("Email sent successfully, status code: %d", response.StatusCode)
	return response.Headers["X-Message-Id"][0], nil
}

func validateSendGridConfig(config *SendGridConfig) error {
	if config.Key == "" || config.From == "" {
		return errors.New("invalid SendGrid configuration")
	}
	return nil
}

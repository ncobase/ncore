package email

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
)

// SMTPConfig holds the configuration for local email sending
type SMTPConfig struct {
	SMTPHost string
	SMTPPort string
	Username string
	Password string
	From     string
}

// LocalSMTPSender implements EmailSender for local SMTP
type LocalSMTPSender struct {
	Config *SMTPConfig
}

func (s *LocalSMTPSender) SendTemplateEmail(recipientEmail string, template Template) (string, error) {
	auth := smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.SMTPHost)
	to := []string{recipientEmail}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipientEmail, template.Subject, fmt.Sprintf("Keyword: %s\nURL: %s", template.Keyword, template.URL)))

	err := smtp.SendMail(fmt.Sprintf("%s:%s", s.Config.SMTPHost, s.Config.SMTPPort), auth, s.Config.From, to, msg)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return "", errors.New("failed to send email")
	}
	log.Printf("Email sent successfully to: %s", recipientEmail)
	return "", nil
}

func validateSMTPConfig(config *SMTPConfig) error {
	if config.SMTPHost == "" || config.SMTPPort == "" || config.Username == "" || config.Password == "" || config.From == "" {
		return errors.New("invalid local email configuration")
	}
	return nil
}

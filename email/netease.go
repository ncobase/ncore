package email

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
)

// NetEaseConfig holds the configuration for NetEase Enterprise Email
type NetEaseConfig struct {
	Username string
	Password string
	From     string
	SMTPHost string
	SMTPPort string
}

// NetEaseSender implements EmailSender for NetEase
type NetEaseSender struct {
	Config *NetEaseConfig
}

func (s *NetEaseSender) SendTemplateEmail(recipientEmail string, template AuthEmailTemplate) (string, error) {
	auth := smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.SMTPHost)
	to := []string{recipientEmail}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipientEmail, template.Subject, fmt.Sprintf("Keyword: %s\nURL: %s", template.Keyword, template.URL)))

	err := smtp.SendMail(fmt.Sprintf("%s:%s", s.Config.SMTPHost, s.Config.SMTPPort), auth, s.Config.From, to, msg)
	if err != nil {
		log.Printf("Error sending email to %s: %v", recipientEmail, err)
		return "", errors.New("failed to send email")
	}
	log.Printf("Email sent successfully to %s", recipientEmail)
	return "", nil
}

func validateNetEaseConfig(config *NetEaseConfig) error {
	if config.Username == "" || config.Password == "" || config.From == "" || config.SMTPHost == "" || config.SMTPPort == "" {
		return errors.New("invalid NetEase configuration")
	}
	return nil
}

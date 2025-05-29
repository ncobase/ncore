package email

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
)

// AliyunConfig holds the configuration for Aliyun DirectMail
type AliyunConfig struct {
	ID      string
	Secret  string
	Account string
}

// AliyunSender implements EmailSender for Aliyun
type AliyunSender struct {
	Config *AliyunConfig
}

func (s *AliyunSender) SendTemplateEmail(recipientEmail string, template Template) (string, error) {
	auth := smtp.PlainAuth("", s.Config.ID, s.Config.Secret, "smtpdm.aliyun.com")
	to := []string{recipientEmail}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipientEmail, template.Subject, fmt.Sprintf("Keyword: %s\nURL: %s", template.Keyword, template.URL)))

	err := smtp.SendMail("smtpdm.aliyun.com:25", auth, s.Config.Account, to, msg)
	if err != nil {
		log.Printf("Error sending email to %s: %v", recipientEmail, err)
		return "", errors.New("failed to send email")
	}
	log.Printf("Email sent successfully to %s", recipientEmail)
	return "", nil
}

func validateAliyunConfig(config *AliyunConfig) error {
	if config.ID == "" || config.Secret == "" || config.Account == "" {
		return errors.New("invalid Aliyun configuration")
	}
	return nil
}

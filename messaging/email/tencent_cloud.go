package email

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
)

// TencentCloudConfig holds the configuration for Tencent Cloud Simple Email Service
type TencentCloudConfig struct {
	ID     string
	Secret string
	From   string
}

// TencentCloudSender implements EmailSender for Tencent Cloud
type TencentCloudSender struct {
	Config *TencentCloudConfig
}

func (s *TencentCloudSender) SendTemplateEmail(recipientEmail string, template AuthEmailTemplate) (string, error) {
	auth := smtp.PlainAuth("", s.Config.ID, s.Config.Secret, "smtp.exmail.qq.com")
	to := []string{recipientEmail}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipientEmail, template.Subject, fmt.Sprintf("Keyword: %s\nURL: %s", template.Keyword, template.URL)))

	err := smtp.SendMail("smtp.exmail.qq.com:25", auth, s.Config.From, to, msg)
	if err != nil {
		log.Printf("Error sending email to %s: %v", recipientEmail, err)
		return "", errors.New("failed to send email")
	}
	log.Printf("Email sent successfully to %s", recipientEmail)
	return "", nil
}

func validateTencentCloudConfig(config *TencentCloudConfig) error {
	if config.ID == "" || config.Secret == "" || config.From == "" {
		return errors.New("invalid Tencent Cloud configuration")
	}
	return nil
}

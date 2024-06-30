package email

import (
	"errors"
)

// Email holds the configuration for all email providers
type Email struct {
	Provider     string
	Mailgun      *MailgunConfig
	Aliyun       *AliyunConfig
	NetEase      *NetEaseConfig
	SendGrid     *SendGridConfig
	SMTP         *SMTPConfig
	TencentCloud *TencentCloudConfig
}

// AuthEmailTemplate represents the email template for authentication
type AuthEmailTemplate struct {
	Subject  string `json:"subject"`
	Template string `json:"template"`
	Keyword  string `json:"keyword"`
	URL      string `json:"url"`
}

// Config is a generic email configuration interface
type Config any

// Sender is a generic interface for sending emails
type Sender interface {
	SendTemplateEmail(recipientEmail string, template AuthEmailTemplate) (string, error)
}

// validateEmailConfig validates the common email configuration
func validateEmailConfig(config Config) error {
	switch c := config.(type) {
	case *MailgunConfig:
		return validateMailgunConfig(c)
	case *AliyunConfig:
		return validateAliyunConfig(c)
	case *NetEaseConfig:
		return validateNetEaseConfig(c)
	case *SendGridConfig:
		return validateSendGridConfig(c)
	case *SMTPConfig:
		return validateSMTPConfig(c)
	case *TencentCloudConfig:
		return validateTencentCloudConfig(c)
	default:
		return errors.New("invalid email configuration")
	}
}

// NewSender returns a new Sender
func NewSender(config Config) (Sender, error) {
	if err := validateEmailConfig(config); err != nil {
		return nil, err
	}
	switch c := config.(type) {
	case *MailgunConfig:
		return &MailgunSender{Config: c}, nil
	case *AliyunConfig:
		return &AliyunSender{Config: c}, nil
	case *NetEaseConfig:
		return &NetEaseSender{Config: c}, nil
	case *SendGridConfig:
		return &SendGridSender{Config: c}, nil
	case *SMTPConfig:
		return &LocalSMTPSender{Config: c}, nil
	case *TencentCloudConfig:
		return &TencentCloudSender{Config: c}, nil
	default:
		return nil, errors.New("create email sender failed")
	}
}

package email

import (
	"errors"
)

// Email holds the configuration for all email providers
type Email struct {
	Provider     string              `json:"provider" yaml:"provider"`
	Mailgun      *MailgunConfig      `json:"mailgun" yaml:"mailgun"`
	Aliyun       *AliyunConfig       `json:"aliyun" yaml:"aliyun"`
	NetEase      *NetEaseConfig      `json:"netease" yaml:"netease"`
	SendGrid     *SendGridConfig     `json:"sendgrid" yaml:"sendgrid"`
	SMTP         *SMTPConfig         `json:"smtp" yaml:"smtp"`
	TencentCloud *TencentCloudConfig `json:"tencent_cloud" yaml:"tencent_cloud"`
}

// Template represents the email template
type Template struct {
	Subject  string `json:"subject"`
	Template string `json:"template"`
	Keyword  string `json:"keyword"`
	URL      string `json:"url"`
	Data     any    `json:"data"`
}

// Config is a generic email configuration interface
type Config any

// Sender is a generic interface for sending emails
type Sender interface {
	SendTemplateEmail(recipientEmail string, template Template) (string, error)
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

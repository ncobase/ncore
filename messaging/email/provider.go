package email

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for the email package.
// It provides email Sender for sending emails through various providers.
//
// Usage:
//
//	wire.Build(
//	    email.ProviderSet,
//	    // ... other providers
//	)
var ProviderSet = wire.NewSet(
	ProvideSender,
	wire.Bind(new(Sender), new(*senderWrapper)),
)

// senderWrapper wraps a Sender interface implementation
type senderWrapper struct {
	sender Sender
}

// SendTemplateEmail delegates to the wrapped sender
func (w *senderWrapper) SendTemplateEmail(recipientEmail string, template Template) (string, error) {
	return w.sender.SendTemplateEmail(recipientEmail, template)
}

// ProvideSender creates an email Sender from Email configuration.
// Returns nil if no valid configuration is provided.
func ProvideSender(cfg *Email) (*senderWrapper, error) {
	if cfg == nil {
		return nil, nil
	}

	var sender Sender
	var err error

	switch cfg.Provider {
	case "mailgun":
		sender, err = NewSender(cfg.Mailgun)
	case "aliyun":
		sender, err = NewSender(cfg.Aliyun)
	case "netease":
		sender, err = NewSender(cfg.NetEase)
	case "sendgrid":
		sender, err = NewSender(cfg.SendGrid)
	case "smtp":
		sender, err = NewSender(cfg.SMTP)
	case "tencent_cloud":
		sender, err = NewSender(cfg.TencentCloud)
	default:
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &senderWrapper{sender: sender}, nil
}

// ProvideMailgunSender creates a Mailgun email sender.
func ProvideMailgunSender(cfg *MailgunConfig) (Sender, error) {
	return NewSender(cfg)
}

// ProvideSendGridSender creates a SendGrid email sender.
func ProvideSendGridSender(cfg *SendGridConfig) (Sender, error) {
	return NewSender(cfg)
}

// ProvideSMTPSender creates an SMTP email sender.
func ProvideSMTPSender(cfg *SMTPConfig) (Sender, error) {
	return NewSender(cfg)
}

package email

import (
	"context"
	"errors"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/ctxutil"
	mail "github.com/ncobase/ncore/messaging/email"
)

const senderKey = "email_sender"

// SetSender stores an email sender in context.
func SetSender(ctx context.Context, sender mail.Sender) context.Context {
	return ctxutil.SetValue(ctx, senderKey, sender)
}

// GetSender returns the configured email sender from context or application config.
func GetSender(ctx context.Context) (mail.Sender, error) {
	if sender, ok := ctxutil.GetValue(ctx, senderKey).(mail.Sender); ok {
		return sender, nil
	}

	cfg := ctxutil.GetConfig(ctx)
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	emailConfig := cfg.Email
	providerConfig, err := ToProviderConfig(emailConfig)
	if err != nil {
		return nil, err
	}

	sender, err := mail.NewSender(providerConfig)
	if err != nil {
		return nil, err
	}

	ctx = SetSender(ctx, sender)
	return sender, nil
}

// ToProviderConfig converts lightweight application config to an email runtime config.
func ToProviderConfig(emailConfig *config.Email) (mail.Config, error) {
	if emailConfig == nil {
		return nil, errors.New("email config is nil")
	}

	switch emailConfig.Provider {
	case "mailgun":
		return toMailgunConfig(emailConfig.Mailgun), nil
	case "aliyun":
		return toAliyunConfig(emailConfig.Aliyun), nil
	case "netease":
		return toNetEaseConfig(emailConfig.NetEase), nil
	case "sendgrid":
		return toSendGridConfig(emailConfig.SendGrid), nil
	case "smtp":
		return toSMTPConfig(emailConfig.SMTP), nil
	case "tencent_cloud":
		return toTencentCloudConfig(emailConfig.TencentCloud), nil
	default:
		return nil, errors.New("unknown email provider")
	}
}

func toMailgunConfig(cfg *config.MailgunConfig) *mail.MailgunConfig {
	if cfg == nil {
		return nil
	}
	return &mail.MailgunConfig{Key: cfg.Key, Domain: cfg.Domain, From: cfg.From}
}

func toAliyunConfig(cfg *config.AliyunConfig) *mail.AliyunConfig {
	if cfg == nil {
		return nil
	}
	return &mail.AliyunConfig{ID: cfg.ID, Secret: cfg.Secret, Account: cfg.Account}
}

func toNetEaseConfig(cfg *config.NetEaseConfig) *mail.NetEaseConfig {
	if cfg == nil {
		return nil
	}
	return &mail.NetEaseConfig{
		Username: cfg.Username,
		Password: cfg.Password,
		From:     cfg.From,
		SMTPHost: cfg.SMTPHost,
		SMTPPort: cfg.SMTPPort,
	}
}

func toSendGridConfig(cfg *config.SendGridConfig) *mail.SendGridConfig {
	if cfg == nil {
		return nil
	}
	return &mail.SendGridConfig{Key: cfg.Key, From: cfg.From}
}

func toSMTPConfig(cfg *config.SMTPConfig) *mail.SMTPConfig {
	if cfg == nil {
		return nil
	}
	return &mail.SMTPConfig{
		SMTPHost: cfg.SMTPHost,
		SMTPPort: cfg.SMTPPort,
		Username: cfg.Username,
		Password: cfg.Password,
		From:     cfg.From,
	}
}

func toTencentCloudConfig(cfg *config.TencentCloudConfig) *mail.TencentCloudConfig {
	if cfg == nil {
		return nil
	}
	return &mail.TencentCloudConfig{ID: cfg.ID, Secret: cfg.Secret, From: cfg.From}
}

// SendWithTemplate sends an email with a template.
func SendWithTemplate(ctx context.Context, recipientEmail string, template mail.Template) (string, error) {
	sender, err := GetSender(ctx)
	if err != nil {
		return "", err
	}
	return sender.SendTemplateEmail(recipientEmail, template)
}

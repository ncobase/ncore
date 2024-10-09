package helper

import (
	"context"
	"errors"

	"ncobase/common/email"
	"ncobase/common/log"
)

// SetEmailSender sets email sender to context.Context
func SetEmailSender(ctx context.Context, sender email.Sender) context.Context {
	return SetValue(ctx, emailSender, sender)
}

// GetEmailSender gets email sender from context.Context based on the configured provider
func GetEmailSender(ctx context.Context) (email.Sender, error) {
	if sender, ok := GetValue(ctx, emailSender).(email.Sender); ok {
		return sender, nil
	}

	// Get email config
	emailConfig := GetConfig(ctx).Email
	var emailProviderConfig email.Config

	// Determine which provider to use based on the configured provider
	switch emailConfig.Provider {
	case "mailgun":
		emailProviderConfig = &emailConfig.Mailgun
	case "aliyun":
		emailProviderConfig = &emailConfig.Aliyun
	case "netease":
		emailProviderConfig = &emailConfig.NetEase
	case "sendgrid":
		emailProviderConfig = &emailConfig.SendGrid
	case "smtp":
		emailProviderConfig = &emailConfig.SMTP
	case "tencent_cloud":
		emailProviderConfig = &emailConfig.TencentCloud
	default:
		return nil, errors.New("unknown email provider")
	}

	// Create email sender based on the configured provider
	sender, err := email.NewSender(emailProviderConfig)
	if err != nil {
		log.Errorf(ctx, "Error creating email sender: %v\n", err)
		return nil, err
	}

	// Set email sender to context.Context for future use
	ctx = SetEmailSender(ctx, sender)
	return sender, nil
}

// SendEmailWithTemplate sends an email with a template
func SendEmailWithTemplate(ctx context.Context, recipientEmail string, template email.AuthEmailTemplate) (string, error) {
	sender, err := GetEmailSender(ctx)
	if err != nil {
		return "", err
	}
	return sender.SendTemplateEmail(recipientEmail, template)
}

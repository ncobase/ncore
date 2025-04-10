package config

import (
	email2 "github.com/ncobase/ncore/pkg/email"

	"github.com/spf13/viper"
)

// Email represents the email configuration
type Email = email2.Email

// getEmailConfig returns the email configuration
func getEmailConfig(v *viper.Viper) *Email {
	return &Email{
		Provider:     v.GetString("email.provider"),
		Mailgun:      getMailgunConfig(v),
		Aliyun:       getAliyunConfig(v),
		NetEase:      getNetEaseConfig(v),
		SendGrid:     getSendGridConfig(v),
		SMTP:         getSMTPConfig(v),
		TencentCloud: getTencentCloudConfig(v),
	}
}
func getMailgunConfig(v *viper.Viper) *email2.MailgunConfig {
	return &email2.MailgunConfig{
		Key:    v.GetString("email.mailgun.key"),
		Domain: v.GetString("email.mailgun.domain"),
		From:   v.GetString("email.mailgun.from"),
	}
}

func getAliyunConfig(v *viper.Viper) *email2.AliyunConfig {
	return &email2.AliyunConfig{
		ID:      v.GetString("email.aliyun.id"),
		Secret:  v.GetString("email.aliyun.secret"),
		Account: v.GetString("email.aliyun.account"),
	}
}

func getNetEaseConfig(v *viper.Viper) *email2.NetEaseConfig {
	return &email2.NetEaseConfig{
		Username: v.GetString("email.netease.username"),
		Password: v.GetString("email.netease.password"),
		From:     v.GetString("email.netease.from"),
		SMTPHost: v.GetString("email.netease.smtp_host"),
		SMTPPort: v.GetString("email.netease.smtp_port"),
	}
}

func getSendGridConfig(v *viper.Viper) *email2.SendGridConfig {
	return &email2.SendGridConfig{
		Key:  v.GetString("email.sendgrid.key"),
		From: v.GetString("email.sendgrid.from"),
	}
}

func getSMTPConfig(v *viper.Viper) *email2.SMTPConfig {
	return &email2.SMTPConfig{
		SMTPHost: v.GetString("email.smtp.host"),
		SMTPPort: v.GetString("email.smtp.port"),
		Username: v.GetString("email.smtp.username"),
		Password: v.GetString("email.smtp.password"),
		From:     v.GetString("email.smtp.from"),
	}
}

func getTencentCloudConfig(v *viper.Viper) *email2.TencentCloudConfig {
	return &email2.TencentCloudConfig{
		ID:     v.GetString("email.tencent_cloud.id"),
		Secret: v.GetString("email.tencent_cloud.secret"),
		From:   v.GetString("email.tencent_cloud.from"),
	}
}

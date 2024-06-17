package config

import (
	"ncobase/common/email"

	"github.com/spf13/viper"
)

// getEmailConfig returns the email configuration
func getEmailConfig(v *viper.Viper) email.Email {
	return email.Email{
		Provider:     v.GetString("email.provider"),
		Mailgun:      getMailgunConfig(v),
		Aliyun:       getAliyunConfig(v),
		NetEase:      getNetEaseConfig(v),
		SendGrid:     getSendGridConfig(v),
		SMTP:         getSMTPConfig(v),
		TencentCloud: getTencentCloudConfig(v),
	}
}
func getMailgunConfig(v *viper.Viper) email.MailgunConfig {
	return email.MailgunConfig{
		Key:    v.GetString("email.mailgun.key"),
		Domain: v.GetString("email.mailgun.domain"),
		From:   v.GetString("email.mailgun.from"),
	}
}

func getAliyunConfig(v *viper.Viper) email.AliyunConfig {
	return email.AliyunConfig{
		ID:      v.GetString("email.aliyun.id"),
		Secret:  v.GetString("email.aliyun.secret"),
		Account: v.GetString("email.aliyun.account"),
	}
}

func getNetEaseConfig(v *viper.Viper) email.NetEaseConfig {
	return email.NetEaseConfig{
		Username: v.GetString("email.netease.username"),
		Password: v.GetString("email.netease.password"),
		From:     v.GetString("email.netease.from"),
		SMTPHost: v.GetString("email.netease.smtp_host"),
		SMTPPort: v.GetString("email.netease.smtp_port"),
	}
}

func getSendGridConfig(v *viper.Viper) email.SendGridConfig {
	return email.SendGridConfig{
		Key:  v.GetString("email.sendgrid.key"),
		From: v.GetString("email.sendgrid.from"),
	}
}

func getSMTPConfig(v *viper.Viper) email.SMTPConfig {
	return email.SMTPConfig{
		SMTPHost: v.GetString("email.smtp.host"),
		SMTPPort: v.GetString("email.smtp.port"),
		Username: v.GetString("email.smtp.username"),
		Password: v.GetString("email.smtp.password"),
		From:     v.GetString("email.smtp.from"),
	}
}

func getTencentCloudConfig(v *viper.Viper) email.TencentCloudConfig {
	return email.TencentCloudConfig{
		ID:     v.GetString("email.tencent_cloud.id"),
		Secret: v.GetString("email.tencent_cloud.secret"),
		From:   v.GetString("email.tencent_cloud.from"),
	}
}

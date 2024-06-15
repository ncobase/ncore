package config

import "ncobase/common/email"

// getEmailConfig returns the email configuration
func getEmailConfig() email.Email {
	return email.Email{
		Provider:     c.GetString("email.provider"),
		Mailgun:      getMailgunConfig(),
		Aliyun:       getAliyunConfig(),
		NetEase:      getNetEaseConfig(),
		SendGrid:     getSendGridConfig(),
		SMTP:         getSMTPConfig(),
		TencentCloud: getTencentCloudConfig(),
	}
}
func getMailgunConfig() email.MailgunConfig {
	return email.MailgunConfig{
		Key:    c.GetString("email.mailgun.key"),
		Domain: c.GetString("email.mailgun.domain"),
		From:   c.GetString("email.mailgun.from"),
	}
}

func getAliyunConfig() email.AliyunConfig {
	return email.AliyunConfig{
		ID:      c.GetString("email.aliyun.id"),
		Secret:  c.GetString("email.aliyun.secret"),
		Account: c.GetString("email.aliyun.account"),
	}
}

func getNetEaseConfig() email.NetEaseConfig {
	return email.NetEaseConfig{
		Username: c.GetString("email.netease.username"),
		Password: c.GetString("email.netease.password"),
		From:     c.GetString("email.netease.from"),
		SMTPHost: c.GetString("email.netease.smtp_host"),
		SMTPPort: c.GetString("email.netease.smtp_port"),
	}
}

func getSendGridConfig() email.SendGridConfig {
	return email.SendGridConfig{
		Key:  c.GetString("email.sendgrid.key"),
		From: c.GetString("email.sendgrid.from"),
	}
}

func getSMTPConfig() email.SMTPConfig {
	return email.SMTPConfig{
		SMTPHost: c.GetString("email.smtp.host"),
		SMTPPort: c.GetString("email.smtp.port"),
		Username: c.GetString("email.smtp.username"),
		Password: c.GetString("email.smtp.password"),
		From:     c.GetString("email.smtp.from"),
	}
}

func getTencentCloudConfig() email.TencentCloudConfig {
	return email.TencentCloudConfig{
		ID:     c.GetString("email.tencent_cloud.id"),
		Secret: c.GetString("email.tencent_cloud.secret"),
		From:   c.GetString("email.tencent_cloud.from"),
	}
}

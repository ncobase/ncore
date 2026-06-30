package config

import (
	"github.com/spf13/viper"
)

// Email contains email provider infrastructure settings.
type Email struct {
	Provider     string              `json:"provider" yaml:"provider"`
	Mailgun      *MailgunConfig      `json:"mailgun" yaml:"mailgun"`
	Aliyun       *AliyunConfig       `json:"aliyun" yaml:"aliyun"`
	NetEase      *NetEaseConfig      `json:"netease" yaml:"netease"`
	SendGrid     *SendGridConfig     `json:"sendgrid" yaml:"sendgrid"`
	SMTP         *SMTPConfig         `json:"smtp" yaml:"smtp"`
	TencentCloud *TencentCloudConfig `json:"tencent_cloud" yaml:"tencent_cloud"`
}

// MailgunConfig contains Mailgun credentials and sender settings.
type MailgunConfig struct {
	Key    string `json:"key" yaml:"key"`
	Domain string `json:"domain" yaml:"domain"`
	From   string `json:"from" yaml:"from"`
}

// AliyunConfig contains Aliyun DirectMail credentials and sender settings.
type AliyunConfig struct {
	ID      string `json:"id" yaml:"id"`
	Secret  string `json:"secret" yaml:"secret"`
	Account string `json:"account" yaml:"account"`
}

// NetEaseConfig contains NetEase SMTP credentials and endpoint settings.
type NetEaseConfig struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	From     string `json:"from" yaml:"from"`
	SMTPHost string `json:"smtp_host" yaml:"smtp_host"`
	SMTPPort string `json:"smtp_port" yaml:"smtp_port"`
}

// SendGridConfig contains SendGrid API credentials and sender settings.
type SendGridConfig struct {
	Key  string `json:"key" yaml:"key"`
	From string `json:"from" yaml:"from"`
}

// SMTPConfig contains generic SMTP credentials and endpoint settings.
type SMTPConfig struct {
	SMTPHost string `json:"host" yaml:"host"`
	SMTPPort string `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	From     string `json:"from" yaml:"from"`
}

// TencentCloudConfig contains Tencent Cloud email credentials and sender settings.
type TencentCloudConfig struct {
	ID     string `json:"id" yaml:"id"`
	Secret string `json:"secret" yaml:"secret"`
	From   string `json:"from" yaml:"from"`
}

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

func getMailgunConfig(v *viper.Viper) *MailgunConfig {
	return &MailgunConfig{
		Key:    v.GetString("email.mailgun.key"),
		Domain: v.GetString("email.mailgun.domain"),
		From:   v.GetString("email.mailgun.from"),
	}
}

func getAliyunConfig(v *viper.Viper) *AliyunConfig {
	return &AliyunConfig{
		ID:      v.GetString("email.aliyun.id"),
		Secret:  v.GetString("email.aliyun.secret"),
		Account: v.GetString("email.aliyun.account"),
	}
}

func getNetEaseConfig(v *viper.Viper) *NetEaseConfig {
	return &NetEaseConfig{
		Username: v.GetString("email.netease.username"),
		Password: v.GetString("email.netease.password"),
		From:     v.GetString("email.netease.from"),
		SMTPHost: v.GetString("email.netease.smtp_host"),
		SMTPPort: v.GetString("email.netease.smtp_port"),
	}
}

func getSendGridConfig(v *viper.Viper) *SendGridConfig {
	return &SendGridConfig{
		Key:  v.GetString("email.sendgrid.key"),
		From: v.GetString("email.sendgrid.from"),
	}
}

func getSMTPConfig(v *viper.Viper) *SMTPConfig {
	return &SMTPConfig{
		SMTPHost: v.GetString("email.smtp.host"),
		SMTPPort: v.GetString("email.smtp.port"),
		Username: v.GetString("email.smtp.username"),
		Password: v.GetString("email.smtp.password"),
		From:     v.GetString("email.smtp.from"),
	}
}

func getTencentCloudConfig(v *viper.Viper) *TencentCloudConfig {
	return &TencentCloudConfig{
		ID:     v.GetString("email.tencent_cloud.id"),
		Secret: v.GetString("email.tencent_cloud.secret"),
		From:   v.GetString("email.tencent_cloud.from"),
	}
}

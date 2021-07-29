package conf

type MailgunConfig struct {
	Key    string
	Domain string
	From   string
}

func getMailgunConfig() *MailgunConfig {
	return &MailgunConfig{
		Key:    c.GetString("mailgun.key"),
		Domain: c.GetString("mailgun.domain"),
		From:   c.GetString("mailgun.from"),
	}
}

package config

// Frontend frontend config struct
type Frontend struct {
	SignInURL string
	SignUpURL string
}

// FrontendConfig returns frontend config
func getFrontendConfig() Frontend {
	return Frontend{
		SignInURL: c.GetString("frontend.sign_in_url"),
		SignUpURL: c.GetString("frontend.sign_up_url"),
	}
}

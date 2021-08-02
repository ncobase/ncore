package conf

// FacebookConfig github config
type FacebookConfig struct {
	ID     string
	Secret string
}

func getFacebookConfig() *FacebookConfig {
	return &FacebookConfig{
		ID:     c.GetString("facebook.id"),
		Secret: c.GetString("facebook.secret"),
	}
}

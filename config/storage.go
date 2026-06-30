package config

import "github.com/spf13/viper"

// Storage contains object storage infrastructure settings.
type Storage struct {
	Provider           string `json:"provider" yaml:"provider"`
	ID                 string `json:"id" yaml:"id"`
	Secret             string `json:"secret" yaml:"secret"`
	Region             string `json:"region" yaml:"region"`
	Bucket             string `json:"bucket" yaml:"bucket"`
	Endpoint           string `json:"endpoint" yaml:"endpoint"`
	ServiceAccountJSON string `json:"service_account_json,omitempty" yaml:"service_account_json,omitempty"`
	SharedFolder       string `json:"shared_folder,omitempty" yaml:"shared_folder,omitempty"`
	OtpCode            string `json:"otp_code,omitempty" yaml:"otp_code,omitempty"`
	Debug              bool   `json:"debug,omitempty" yaml:"debug,omitempty"`
	AppID              string `json:"app_id,omitempty" yaml:"app_id,omitempty"`
}

func getStorageConfig(v *viper.Viper) *Storage {
	return &Storage{
		Provider:           v.GetString("storage.provider"),
		ID:                 v.GetString("storage.id"),
		Secret:             v.GetString("storage.secret"),
		Region:             v.GetString("storage.region"),
		Bucket:             v.GetString("storage.bucket"),
		Endpoint:           v.GetString("storage.endpoint"),
		ServiceAccountJSON: v.GetString("storage.service_account_json"),
		SharedFolder:       v.GetString("storage.shared_folder"),
		OtpCode:            v.GetString("storage.otp_code"),
		Debug:              v.GetBool("storage.debug"),
		AppID:              v.GetString("storage.app_id"),
	}
}

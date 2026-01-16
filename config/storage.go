package config

import (
	"github.com/ncobase/ncore/oss"
	"github.com/spf13/viper"
)

type Storage = oss.Config

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

package config

import "github.com/spf13/viper"

// Desensitization holds desensitization settings
type Desensitization struct {
	Enabled               bool     `json:"enabled" yaml:"enabled"`
	SensitiveFields       []string `json:"sensitive_fields" yaml:"sensitive_fields"`
	CustomPatterns        []string `json:"custom_patterns" yaml:"custom_patterns"`
	PreservePrefix        int      `json:"preserve_prefix" yaml:"preserve_prefix"`
	PreserveSuffix        int      `json:"preserve_suffix" yaml:"preserve_suffix"`
	MaskChar              string   `json:"mask_char" yaml:"mask_char"`
	UseFixedLength        bool     `json:"use_fixed_length" yaml:"use_fixed_length"`
	FixedMaskLength       int      `json:"fixed_mask_length" yaml:"fixed_mask_length"`
	ExactFieldMatch       bool     `json:"exact_field_match" yaml:"exact_field_match"`
	EnableDefaultPatterns bool     `json:"enable_default_patterns" yaml:"enable_default_patterns"`
}

// Default sensitive field patterns
var defaultSensitiveFields = []string{
	"password", "passwd", "pwd",
	"token", "access_token", "refresh_token", "auth_token",
	"secret", "api_key", "apikey",
	"credit_card", "card_number",
}

// Default mask character
const defaultMaskChar = "*"

// Default fixed mask length
const defaultFixedMaskLength = 6

// Enable exact field match
const enableExactFieldMatch = false

// Enable default patterns
const enableDefaultPatterns = false

// getDesensitizationConfigs reads and returns desensitization configuration
func getDesensitizationConfigs(v *viper.Viper) *Desensitization {
	if !v.IsSet("logger.desensitization") {
		return &Desensitization{
			Enabled:               true,
			SensitiveFields:       defaultSensitiveFields,
			PreservePrefix:        0,
			PreserveSuffix:        0,
			MaskChar:              defaultMaskChar,
			UseFixedLength:        true,
			FixedMaskLength:       defaultFixedMaskLength,
			ExactFieldMatch:       enableExactFieldMatch,
			EnableDefaultPatterns: enableDefaultPatterns,
		}
	}

	config := &Desensitization{
		Enabled:               v.GetBool("logger.desensitization.enabled"),
		SensitiveFields:       v.GetStringSlice("logger.desensitization.sensitive_fields"),
		CustomPatterns:        v.GetStringSlice("logger.desensitization.custom_patterns"),
		PreservePrefix:        v.GetInt("logger.desensitization.preserve_prefix"),
		PreserveSuffix:        v.GetInt("logger.desensitization.preserve_suffix"),
		MaskChar:              v.GetString("logger.desensitization.mask_char"),
		UseFixedLength:        v.GetBool("logger.desensitization.use_fixed_length"),
		FixedMaskLength:       v.GetInt("logger.desensitization.fixed_mask_length"),
		ExactFieldMatch:       v.GetBool("logger.desensitization.exact_field_match"),
		EnableDefaultPatterns: v.GetBool("logger.desensitization.enable_default_patterns"),
	}

	// Apply defaults for missing values
	if len(config.SensitiveFields) == 0 {
		config.SensitiveFields = defaultSensitiveFields
	}
	if config.MaskChar == "" {
		config.MaskChar = "*"
	}
	if config.FixedMaskLength == 0 {
		config.FixedMaskLength = 8
	}
	if !v.IsSet("logger.desensitization.exact_field_match") {
		config.ExactFieldMatch = enableExactFieldMatch
	}
	if !v.IsSet("logger.desensitization.enable_default_patterns") {
		config.EnableDefaultPatterns = enableDefaultPatterns
	}

	return config
}

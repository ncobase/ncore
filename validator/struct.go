package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// errorMessages is a nested map of languages to validation tags to custom error messages.
var errorMessages = map[string]map[string]string{
	"en": {
		"required": "The field '%s' is required.",
		"email":    "The field '%s' must be a valid email address.",
		"min":      "The field '%s' must be at least %s characters long.",
		"max":      "The field '%s' must be no longer than %s characters.",
		"lte":      "The field '%s' must be less than or equal to %s.",
		"gte":      "The field '%s' must be greater than or equal to %s.",
		"unique":   "The field '%s' must be unique.",
		"gt":       "The field '%s' must be greater than %s.",
		"lt":       "The field '%s' must be less than %s.",
		"enum":     "The field '%s' must be one of %s.",
	},
	"zh": {
		"required": "字段 '%s' 为必填项。",
		"email":    "字段 '%s' 必须是有效的电子邮箱地址。",
		"min":      "字段 '%s' 的长度不能少于 %s 个字符。",
		"max":      "字段 '%s' 的长度不能超过 %s 个字符。",
		"lte":      "字段 '%s' 的值必须小于或等于 %s。",
		"gte":      "字段 '%s' 的值必须大于或等于 %s。",
		"unique":   "字段 '%s' 的值必须唯一。",
		"gt":       "字段 '%s' 的值必须大于 %s。",
		"lt":       "字段 '%s' 的值必须小于 %s。",
		"enum":     "字段 '%s' 的值必须是 %s 之一。",
	},
	// Add more languages as needed.
}

// parseMessage constructs a friendly error message based on the validation tag and custom messages.
func parseMessage(jsonTag string, e validator.FieldError, lang ...string) string {
	var msgLang string
	if len(lang) > 0 {
		msgLang = lang[0]
	} else {
		msgLang = "en"
	}
	if msgs, exists := errorMessages[msgLang]; exists {
		if msg, exists := msgs[e.Tag()]; exists {
			// Check the number of %s placeholders in the custom message
			placeholderCount := strings.Count(msg, "%s")
			if placeholderCount == 1 {
				return fmt.Sprintf(msg, jsonTag)
			} else if placeholderCount == 2 {
				return fmt.Sprintf(msg, jsonTag, e.Param())
			}
		}
	}
	// Default error message if no custom message is defined for the tag or language.
	return fmt.Sprintf("Field '%s' is invalid: %s", jsonTag, e.Tag())
}

// ValidateStruct validates a struct and returns a map of JSON field names to friendly error messages.
func ValidateStruct(s any, lang ...string) map[string]string {
	validationErrors := make(map[string]string)

	err := validate.Struct(s)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			structType := reflect.TypeOf(s).Elem()
			for _, e := range validationErrs {
				field, _ := structType.FieldByName(e.StructField())
				jsonTag := field.Tag.Get("json")
				if jsonTag == "" {
					jsonTag = e.StructField()
				} else {
					jsonTag = strings.Split(jsonTag, ",")[0]
				}
				validationErrors[jsonTag] = parseMessage(jsonTag, e, lang...)
			}
		}
	}

	return validationErrors
}

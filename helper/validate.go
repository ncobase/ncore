package helper

import (
	"ncobase/ncore/validator"

	"github.com/gin-gonic/gin"
)

// Validate is a wrapper around validator.Validate that returns a map of JSON field names to friendly error messages.
var Validate = validator.ValidateStruct

// ShouldBindAndValidateStruct binds and validates struct
func ShouldBindAndValidateStruct(c *gin.Context, obj any, lang ...string) (map[string]string, error) {
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/json;charset=utf-8"
	}

	if err := c.ShouldBind(obj); err != nil {
		return nil, err
	}

	return Validate(obj, lang...), nil
}

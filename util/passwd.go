package util

import (
	"context"
	"stone/common/log"

	"golang.org/x/crypto/bcrypt"
)

// EncryptPassword - Encrypt password
func EncryptPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf(context.Background(), "util.EncryptPassword error: %v", err.Error())
		return err.Error()
	}
	return string(hash)
}

// ComparePassword - Verify encrypt password the consistent
func ComparePassword(encodePassword, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(encodePassword), []byte(password)); err != nil {
		return false
	}
	return true
}

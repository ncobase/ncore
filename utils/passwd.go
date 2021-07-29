package utils

import (
	"context"
	"stone/common/log"

	"golang.org/x/crypto/bcrypt"
)

// EncryptPassword 对密码进行加密
func EncryptPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf(context.Background(), "utils.EncryptPassword error: %v\n", err.Error())
		return err.Error()
	}
	return string(hash)
}

// ComparePassword 验证加密密码
func ComparePassword(encodePassword, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(encodePassword), []byte(password)); err != nil {
		return false
	}
	return true
}

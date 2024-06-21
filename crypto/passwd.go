package crypto

import (
	"context"

	"ncobase/common/log"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the provided password using bcrypt.
func HashPassword(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf(ctx, "encrypt.HashPassword error: %v", err)
		return "", err
	}
	return string(hash), nil
}

// ComparePassword compares the hashed password with the provided password.
func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

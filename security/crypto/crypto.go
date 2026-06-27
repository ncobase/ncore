package crypto

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/ncobase/ncore/logging/logger"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = bcrypt.DefaultCost
)

var (
	ErrInvalidPadding = errors.New("invalid PKCS7 padding")
)

// HashPassword hashes the provided password using bcrypt.
func HashPassword(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		logger.Errorf(ctx, "encrypt.HashPassword error: %v", err)
		return "", err
	}
	return string(hash), nil
}

// ComparePassword compares the hashed password with the provided password.
func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// PKCS7Padding - PKCS7 padding
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	if blockSize <= 0 {
		return ciphertext
	}
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding - PKCS7 unpadding
func PKCS7UnPadding(origData []byte) []byte {
	data, err := PKCS7UnPaddingE(origData)
	if err != nil {
		return nil
	}
	return data
}

// PKCS7UnPaddingE removes PKCS7 padding and validates the padding bytes.
func PKCS7UnPaddingE(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, ErrInvalidPadding
	}
	unpadding := int(origData[length-1])
	if unpadding == 0 || unpadding > length {
		return nil, ErrInvalidPadding
	}
	for _, b := range origData[length-unpadding:] {
		if int(b) != unpadding {
			return nil, ErrInvalidPadding
		}
	}
	return origData[:length-unpadding], nil
}

// AesEncrypt encrypts plaintext with AES-CFB. Prefer AesGCMEncrypt for new
// code because CFB does not authenticate ciphertext integrity.
func AesEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	copy(ciphertext[:aes.BlockSize], iv)

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// AesDecrypt decrypts AES-CFB ciphertext produced by AesEncrypt.
func AesDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	payload := make([]byte, len(ciphertext)-aes.BlockSize)
	copy(payload, ciphertext[aes.BlockSize:])

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(payload, payload)

	return payload, nil
}

// AesGCMEncrypt encrypts and authenticates plaintext with AES-GCM. The returned
// byte slice is nonce || ciphertext || tag.
func AesGCMEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	result := make([]byte, 0, len(nonce)+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)
	return result, nil
}

// AesGCMDecrypt decrypts and authenticates data produced by AesGCMEncrypt.
func AesGCMDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	payload := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, payload, nil)
}

// AesEncrypt64 - AES Encrypt base64
func AesEncrypt64(plaintext []byte, key []byte) (string, error) {
	ciphertext, err := AesEncrypt(plaintext, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AesDecrypt64 - AES Decrypt base64
func AesDecrypt64(encodedCiphertext string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}
	return AesDecrypt(ciphertext, key)
}

// AesGCMEncrypt64 encrypts and authenticates plaintext, returning base64 data.
func AesGCMEncrypt64(plaintext []byte, key []byte) (string, error) {
	ciphertext, err := AesGCMEncrypt(plaintext, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AesGCMDecrypt64 decrypts base64 data produced by AesGCMEncrypt64.
func AesGCMDecrypt64(encodedCiphertext string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}
	return AesGCMDecrypt(ciphertext, key)
}

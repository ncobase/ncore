package crypto

import (
	"bytes"
	"errors"
	"testing"
)

func TestPKCS7UnPaddingEValidatesPadding(t *testing.T) {
	padded := PKCS7Padding([]byte("payload"), 16)
	unpadded, err := PKCS7UnPaddingE(padded)
	if err != nil {
		t.Fatalf("unpadding valid data: %v", err)
	}
	if string(unpadded) != "payload" {
		t.Fatalf("unexpected unpadded data: %q", string(unpadded))
	}

	if _, err := PKCS7UnPaddingE([]byte{1, 2, 3, 4}); !errors.Is(err, ErrInvalidPadding) {
		t.Fatalf("expected ErrInvalidPadding, got %v", err)
	}
}

func TestAesGCMEncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte("sensitive payload")

	ciphertext, err := AesGCMEncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Contains(ciphertext, plaintext) {
		t.Fatal("ciphertext contains plaintext")
	}

	decrypted, err := AesGCMDecrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("unexpected decrypted payload: %q", string(decrypted))
	}

	ciphertext[len(ciphertext)-1] ^= 0x01
	if _, err := AesGCMDecrypt(ciphertext, key); err == nil {
		t.Fatal("expected authentication error for tampered ciphertext")
	}
}

func TestAesCFBDecryptDoesNotMutateInput(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	ciphertext, err := AesEncrypt([]byte("payload"), key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	original := append([]byte(nil), ciphertext...)

	if _, err := AesDecrypt(ciphertext, key); err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(ciphertext, original) {
		t.Fatal("AesDecrypt mutated input ciphertext")
	}
}

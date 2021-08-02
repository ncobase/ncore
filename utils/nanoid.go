package utils

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	alphabet string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	length   int    = 22
)

// DefaultNanoID Generate default NanoID
func DefaultNanoID(l ...int) string {
	return gonanoid.Must(l...)
}

// NanoString Nano String
func NanoString(len int) string {
	return gonanoid.MustGenerate(alphabet, len)
}

// NanoID Nano ID
func NanoID() string {
	return gonanoid.MustGenerate(alphabet, length)
}

// GenerateSlug default length 11
func GenerateSlug(l ...int) string {
	dl := 11
	if l != nil {
		dl = l[0]
	}
	return DefaultNanoID(dl)
}

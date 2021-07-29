package utils

import (
	"github.com/matoous/go-nanoid/v2"
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

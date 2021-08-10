package util

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	defaultAlphabet = "23456789ABCDEFGHJKLMNPQRSTWXYZabcdefghijkmnopqrstuvwxyz"
	defaultSize     = 16
)

// NanoID Generate optional length nanoid, use const by default
func NanoID(l ...int) string {
	var size int
	switch {
	case len(l) == 0:
		size = defaultSize
	case len(l) == 1:
		size = l[0]
	}
	return gonanoid.MustGenerate(defaultAlphabet, size)
}

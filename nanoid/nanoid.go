package nanoid

import (
	"strings"

	"github.com/ncobase/common/consts"
	"github.com/ncobase/common/validator"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	// alphabet = "23456789ABCDEFGHJKLMNPQRSTWXYZabcdefghijkmnopqrstuvwxyz" // remove resemble letters
	alphabet       = consts.Number + consts.Lowercase + consts.Uppercase
	defaultSize    = 11
	PrimaryKeySize = defaultSize
)

var DefaultAlphabetLen = len(alphabet)

func getSize(l ...int) int {
	size := defaultSize
	if len(l) > 0 {
		size = l[0]
	}
	return size
}

// Must generate optional length nanoid
func Must(l ...int) string {
	size := getSize(l...)
	return gonanoid.Must(size)
}

// String generate optional length nanoid, use const by default
func String(l ...int) string {
	size := getSize(l...)
	return gonanoid.MustGenerate(alphabet, size)
}

// Lower generate optional length nanoid, use const by default
func Lower(l ...int) string {
	size := getSize(l...)
	return gonanoid.MustGenerate(consts.Lowercase, size)
}

// Upper generate optional length nanoid, use const by default
func Upper(l ...int) string {
	size := getSize(l...)
	return gonanoid.MustGenerate(consts.Uppercase, size)
}

// Number generate optional length nanoid, use const by default
func Number(l ...int) string {
	size := getSize(l...)
	return gonanoid.MustGenerate(consts.Number, size)
}

// PrimaryKey generate primary key
func PrimaryKey(l ...int) func() string {
	size := PrimaryKeySize
	if len(l) > 0 {
		size = l[0]
	}
	return func() string {
		return gonanoid.MustGenerate(alphabet, size)
	}
}

// IsPrimaryKey verify is primary key
func IsPrimaryKey(id string) bool {
	if validator.IsEmpty(id) {
		return false
	}
	size := PrimaryKeySize
	strLen := len(id)
	inAlphabet := strings.ContainsAny(alphabet, id) && (DefaultAlphabetLen*size >= strLen*4)
	return strLen == size && inAlphabet
}

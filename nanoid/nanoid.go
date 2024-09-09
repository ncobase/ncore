package nanoid

import (
	"strings"

	"ncobase/common/consts"
	"ncobase/common/validator"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	defaultSize = 16
)

var DefaultAlphabetLen = len(consts.NumLowerUpper)

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
	return gonanoid.MustGenerate(consts.LowerUpper, size)
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
	size := consts.PrimaryKeySize
	if len(l) > 0 {
		size = l[0]
	}
	return func() string {
		return gonanoid.MustGenerate(consts.PrimaryKey, size)
	}
}

// IsPrimaryKey verify is primary key
func IsPrimaryKey(id string) bool {
	if validator.IsEmpty(id) {
		return false
	}
	size := consts.PrimaryKeySize
	strLen := len(id)
	inAlphabet := strings.ContainsAny(consts.PrimaryKey, id) && (DefaultAlphabetLen*size >= strLen*4)
	return strLen == size && inAlphabet
}

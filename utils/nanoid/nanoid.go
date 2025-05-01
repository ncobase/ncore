package nanoid

import (
	"crypto/rand"
	"math/big"
	"strings"
	"sync"

	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/validation/validator"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	defaultSize = 16
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, consts.PrimaryKeySize)
		},
	}
)

// getSize returns the provided size or the default size if not provided
func getSize(l ...int) int {
	if len(l) > 0 {
		return l[0]
	}
	return defaultSize
}

// Must generates a NanoID with optional length using default alphabet
func Must(l ...int) string {
	size := getSize(l...)
	return gonanoid.Must(size)
}

// String generates a NanoID using only letters with optional length
func String(l ...int) string {
	size := getSize(l...)
	return gonanoid.MustGenerate(consts.LowerUpper, size)
}

// Lower generates a NanoID using only lowercase letters with optional length
func Lower(l ...int) string {
	size := getSize(l...)
	return gonanoid.MustGenerate(consts.Lowercase, size)
}

// Upper generates a NanoID using only uppercase letters with optional length
func Upper(l ...int) string {
	size := getSize(l...)
	return gonanoid.MustGenerate(consts.Uppercase, size)
}

// Number generates a NanoID using only numbers with optional length
func Number(l ...int) string {
	size := getSize(l...)
	return gonanoid.MustGenerate(consts.Number, size)
}

// PrimaryKey returns a function that generates primary keys with specified length
func PrimaryKey(l ...int) func() string {
	size := consts.PrimaryKeySize
	if len(l) > 0 {
		size = l[0]
	}
	return func() string {
		return generateFastID(size, consts.PrimaryKey)
	}
}

// IsPrimaryKey verifies if a string is a valid primary key
func IsPrimaryKey(id string) bool {
	if validator.IsEmpty(id) {
		return false
	}
	size := consts.PrimaryKeySize
	strLen := len(id)

	if strLen != size {
		return false
	}

	for _, c := range id {
		if !strings.ContainsRune(consts.PrimaryKey, c) {
			return false
		}
	}
	return true
}

// generateFastID creates an ID with specified size and alphabet using optimized algorithm
func generateFastID(size int, alphabet string) string {
	alphabetLen := big.NewInt(int64(len(alphabet)))
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	result := make([]byte, size)

	_, err := rand.Read(buf[:size])
	if err != nil {
		return gonanoid.MustGenerate(alphabet, size)
	}

	for i := 0; i < size; i++ {
		n := new(big.Int).SetInt64(int64(buf[i]))
		n.Mod(n, alphabetLen)
		result[i] = alphabet[n.Int64()]
	}

	return string(result)
}

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"stone/common/conf"
)

// CreateHash 해시를 생성하는 함수
func CreateHash(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))

	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)

	return mdStr
}

// GetEnvWithKey process.env 에 key값과 일치하는 값을 가져온다
func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}

// Find returns the smallest index i at which x == a[i],
// or len(a) if there is no such index.
func Find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

func FindID(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// GetDomain Get the run domain
func GetDomain(domain string) string {
	if domain == "localhost" {
		return fmt.Sprintf("%v://%v:%d", conf.G.Protocol, conf.G.Domain, conf.G.Port)
	}
	return fmt.Sprintf("%v://%v", conf.G.Protocol, conf.G.Domain)
}

package ctxutil

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// generateAntifakeCode generate antifake code
func generateAntifakeCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based code if crypto/rand fails
		// This is extremely rare but provides resilience
		timestamp := time.Now().UnixNano()
		for i := 0; i < length; i++ {
			bytes[i] = charset[timestamp%int64(len(charset))]
			timestamp /= int64(len(charset))
		}
		return string(bytes)
	}
	for i := 0; i < length; i++ {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(bytes)
}

var (
	mu           sync.Mutex
	currentMonth string
	counters     = make(map[string]int)
)

// GenerateBusinessCode generate business code
//
// format: Code+YYYYMM+AntifakeCODE+Serial, eg: CO2009110GA0001
func GenerateBusinessCode(identifier string) string {
	currentDate := time.Now().Format("200601")

	mu.Lock()
	defer mu.Unlock()

	if currentMonth != currentDate {
		currentMonth = currentDate
		counters = make(map[string]int)
	}

	counters[identifier]++
	serialNumber := counters[identifier]

	antifakeCode := generateAntifakeCode(3)

	businessCode := fmt.Sprintf("%s%s%s%04d", identifier, currentDate, antifakeCode, serialNumber)

	return businessCode
}

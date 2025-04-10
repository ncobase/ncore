package types

import "time"

// EventData basic event data
type EventData struct {
	Time      time.Time
	Source    string
	EventType string
	Data      any
}

package types

import (
	"errors"
	"time"
)

// Event data structure
type EventData struct {
	Time      time.Time
	Source    string
	EventType string
	Data      any
}

// Extract payload from event data
func ExtractEventPayload(data any) (*map[string]any, error) {
	eventData, ok := data.(EventData)
	if !ok {
		return nil, errors.New("invalid event data format")
	}

	if payload, ok := eventData.Data.(*map[string]any); ok {
		return payload, nil
	}
	return nil, errors.New("invalid payload format")
}

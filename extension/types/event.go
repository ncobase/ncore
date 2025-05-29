package types

import (
	"encoding/json"
	"errors"
	"time"
)

// EventTarget defines where an event should be published
type EventTarget int

const (
	EventTargetMemory EventTarget                            = 1 << iota // In-memory event bus
	EventTargetQueue                                                     // Message queue (RabbitMQ/Kafka)
	EventTargetAll    = EventTargetMemory | EventTargetQueue             // All available targets
)

// EventData Event data structure
type EventData struct {
	Time      time.Time `json:"time"`
	Source    string    `json:"source"`
	EventType string    `json:"event_type"`
	Data      any       `json:"data"`
}

// ExtractEventPayload Extract payload from event data
func ExtractEventPayload(data any) (*map[string]any, error) {
	// If data is nil, return empty map
	if data == nil {
		emptyMap := make(map[string]any)
		return &emptyMap, nil
	}

	// If it's already an EventData struct
	if eventData, ok := data.(EventData); ok {
		return extractFromData(eventData.Data)
	}

	// If it's a pointer to EventData struct
	if eventDataPtr, ok := data.(*EventData); ok && eventDataPtr != nil {
		return extractFromData(eventDataPtr.Data)
	}

	// Try to process as a raw map
	return extractFromData(data)
}

// extractFromData extracts payload from raw data formats
func extractFromData(data any) (*map[string]any, error) {
	// Handle nil case
	if data == nil {
		emptyMap := make(map[string]any)
		return &emptyMap, nil
	}

	// Already a map pointer
	if mapPtr, ok := data.(*map[string]any); ok {
		if mapPtr == nil {
			emptyMap := make(map[string]any)
			return &emptyMap, nil
		}
		return mapPtr, nil
	}

	// Already a map
	if m, ok := data.(map[string]any); ok {
		return &m, nil
	}

	// Handle string (JSON) data - common when receiving from message queue
	if strData, ok := data.(string); ok {
		var result map[string]any
		if err := json.Unmarshal([]byte(strData), &result); err != nil {
			return nil, errors.New("failed to unmarshal string payload: " + err.Error())
		}
		return &result, nil
	}

	// Handle []byte (JSON) data - also common from MQ
	if byteData, ok := data.([]byte); ok {
		var result map[string]any
		if err := json.Unmarshal(byteData, &result); err != nil {
			return nil, errors.New("failed to unmarshal byte payload: " + err.Error())
		}
		return &result, nil
	}

	// Try JSON marshaling and unmarshaling for other types
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New("invalid payload format, cannot marshal: " + err.Error())
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, errors.New("invalid payload format, cannot unmarshal: " + err.Error())
	}

	return &result, nil
}

// SafeGet safely extracts a value from a payload with type assertion
// Usage: value := SafeGet[string](payload, "key")
// Returns zero value of T if key doesn't exist or value can't be converted to T
func SafeGet[T any](payload *map[string]any, key string) T {
	var zero T
	if payload == nil {
		return zero
	}

	value, exists := (*payload)[key]
	if !exists || value == nil {
		return zero
	}

	// Try type assertion
	typed, ok := value.(T)
	if !ok {
		return zero
	}

	return typed
}

// SafeGetWithDefault safely extracts a value with a default fallback
// Usage: value := SafeGetWithDefault(payload, "key", defaultValue)
func SafeGetWithDefault[T any](payload *map[string]any, key string, defaultValue T) T {
	if payload == nil {
		return defaultValue
	}

	value, exists := (*payload)[key]
	if !exists || value == nil {
		return defaultValue
	}

	// Try type assertion
	typed, ok := value.(T)
	if !ok {
		return defaultValue
	}

	return typed
}

// SafeGetOr safely extracts a value with a provided function for handling missing/nil values
// Usage: value := SafeGetOr(payload, "key", func() T { return computedDefault })
func SafeGetOr[T any](payload *map[string]any, key string, orElse func() T) T {
	if payload == nil {
		return orElse()
	}

	value, exists := (*payload)[key]
	if !exists || value == nil {
		return orElse()
	}

	// Try type assertion
	typed, ok := value.(T)
	if !ok {
		return orElse()
	}

	return typed
}

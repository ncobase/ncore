package convert

import (
	"encoding/json"
	"fmt"
)

// ToJSON serializes any object to JSON string
func ToJSON(v any) (string, error) {
	if v == nil {
		return "", nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(data), nil
}

// FromJSON deserializes JSON string to specified object
func FromJSON(s string, v any) error {
	if s == "" {
		return nil
	}

	err := json.Unmarshal([]byte(s), v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal from JSON: %w", err)
	}

	return nil
}

// MustToJSON serializes any object to JSON string, panics on error
func MustToJSON(v any) string {
	s, err := ToJSON(v)
	if err != nil {
		panic(err)
	}
	return s
}

// MustFromJSON deserializes JSON string to specified object, panics on error
func MustFromJSON(s string, v any) {
	err := FromJSON(s, v)
	if err != nil {
		panic(err)
	}
}

// ToJSONMap serializes any object to map[string]any
func ToJSONMap(v any) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}

	// Serialize to JSON first
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	// Then deserialize to map
	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return result, nil
}

// FromJSONMap deserializes map[string]any to specified object
func FromJSONMap(m map[string]any, v any) error {
	if m == nil {
		return nil
	}

	// Serialize map to JSON first
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	// Then deserialize to target object
	err = json.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal from map: %w", err)
	}

	return nil
}

// PrettyJSON formats object as indented JSON string
func PrettyJSON(v any) (string, error) {
	if v == nil {
		return "", nil
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to pretty JSON: %w", err)
	}

	return string(data), nil
}

// StringifyJSON converts object to JSON string, returns empty string on error
func StringifyJSON(v any) string {
	s, err := ToJSON(v)
	if err != nil {
		return ""
	}
	return s
}

// ParseJSON parses JSON string to specified object, returns false on error
func ParseJSON(s string, v any) bool {
	err := FromJSON(s, v)
	return err == nil
}

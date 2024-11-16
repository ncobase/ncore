package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

// ConvertToSyncMap takes a regular map[string]T and returns a pointer to a sync.Map.
// The function copies all the key-value pairs from the regular map into the sync.Map.
func ConvertToSyncMap[T any](m map[string]T) *sync.Map {
	sm := &sync.Map{}
	for key, value := range m {
		sm.Store(key, value)
	}
	return sm
}

// ConvertToMap takes a sync.Map and converts it into a regular map[string]T.
// The function iterates over the sync.Map and retrieves each key-value pair,
// returning them as a standard Go map.
func ConvertToMap[T any](sm *sync.Map) map[string]T {
	m := make(map[string]T)
	sm.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			if v, ok := value.(T); ok {
				m[k] = v
			}
		}
		return true
	})
	return m
}

// TypeConverter handles the conversion of values between different types
type TypeConverter struct{}

// ConvertValue converts a string value to the corresponding type based on the type string
// Supported types: string, number, boolean, object, array
func ConvertValue(typeStr string, value string) (any, error) {
	switch typeStr {
	case "string":
		return value, nil

	case "number":
		// Try to convert to float64 first
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			// Check if it's actually an integer
			if float64(int64(floatVal)) == floatVal {
				return int64(floatVal), nil
			}
			return floatVal, nil
		}
		return nil, fmt.Errorf("invalid number format: %s", value)

	case "boolean":
		switch value {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid boolean value: %s", value)
		}

	case "object":
		var result map[string]any
		if err := json.Unmarshal([]byte(value), &result); err != nil {
			return nil, fmt.Errorf("invalid object format: %v", err)
		}
		return result, nil

	case "array":
		var result []any
		if err := json.Unmarshal([]byte(value), &result); err != nil {
			return nil, fmt.Errorf("invalid array format: %v", err)
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", typeStr)
	}
}

// ToString converts any value to its string representation
// Returns empty string for nil values
func ToString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case map[string]any, []any:
		if bytes, err := json.Marshal(v); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToInt attempts to convert a value to an integer
// Supports conversion from string, float64, bool, and integer types
func ToInt(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// ToFloat attempts to convert a value to a float64
// Supports conversion from string, integer types, and bool
func ToFloat(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float", value)
	}
}

// ToBool attempts to convert a value to a boolean
// Supports conversion from string, number types, and bool
// For strings, accepts "true", "1", "yes", "on" as true values
// and "false", "0", "no", "off", "" as false values
func ToBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		switch v {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off", "":
			return false, nil
		default:
			return false, fmt.Errorf("invalid boolean value: %s", v)
		}
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// ToObject attempts to convert a value to a map[string]interface{}
// Supports conversion from string (JSON) and existing maps
func ToObject(value any) (map[string]any, error) {
	switch v := value.(type) {
	case map[string]any:
		return v, nil
	case string:
		var result map[string]any
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			return nil, fmt.Errorf("invalid object format: %v", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to object", value)
	}
}

// ToArray attempts to convert a value to []interface{}
// Supports conversion from string (JSON) and existing arrays
func ToArray(value any) ([]any, error) {
	switch v := value.(type) {
	case []any:
		return v, nil
	case string:
		var result []any
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			return nil, fmt.Errorf("invalid array format: %v", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to array", value)
	}
}

// IsValidType checks if the provided type string is supported
// Returns true for valid types: string, number, boolean, object, array
func IsValidType(typeStr string) bool {
	validTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"boolean": true,
		"object":  true,
		"array":   true,
	}
	return validTypes[typeStr]
}

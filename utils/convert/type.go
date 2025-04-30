package convert

import (
	"encoding/json"
	"fmt"
	"strconv"
)

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

// ToStringArray converts any type of array to a string array
// Each element is converted to its string representation
func ToStringArray(arr any) []string {
	if arr == nil {
		return nil
	}

	// Use reflection to handle different array types
	switch v := arr.(type) {
	case []string:
		// Already a string array, return a copy
		result := make([]string, len(v))
		copy(result, v)
		return result

	case []int:
		result := make([]string, len(v))
		for i, val := range v {
			result[i] = IntToString(val)
		}
		return result

	case []int64:
		result := make([]string, len(v))
		for i, val := range v {
			result[i] = Int64ToString(val)
		}
		return result

	case []float64:
		result := make([]string, len(v))
		for i, val := range v {
			result[i] = strconv.FormatFloat(val, 'f', -1, 64)
		}
		return result

	case []bool:
		result := make([]string, len(v))
		for i, val := range v {
			result[i] = strconv.FormatBool(val)
		}
		return result

	case []any:
		result := make([]string, len(v))
		for i, val := range v {
			result[i] = ToString(val)
		}
		return result

	default:
		// Try to handle it as a slice of interface{}
		if data, err := json.Marshal(v); err == nil {
			var anySlice []any
			if json.Unmarshal(data, &anySlice) == nil {
				return ToStringArray(anySlice)
			}
		}
		return []string{}
	}
}

// FromStringArray converts a string array to an array of a specific type
// The conversion is based on the target type
func FromStringArray(arr []string, targetType string) (any, error) {
	if arr == nil {
		return nil, nil
	}

	switch targetType {
	case "string":
		// Already a string array, return a copy
		result := make([]string, len(arr))
		copy(result, arr)
		return result, nil

	case "int":
		result := make([]int, len(arr))
		for i, val := range arr {
			intVal, err := StringToInt(val)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element at index %d to int: %w", i, err)
			}
			result[i] = intVal
		}
		return result, nil

	case "int64":
		result := make([]int64, len(arr))
		for i, val := range arr {
			int64Val, err := StringToInt64(val)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element at index %d to int64: %w", i, err)
			}
			result[i] = int64Val
		}
		return result, nil

	case "float64":
		result := make([]float64, len(arr))
		for i, val := range arr {
			floatVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element at index %d to float64: %w", i, err)
			}
			result[i] = floatVal
		}
		return result, nil

	case "bool":
		result := make([]bool, len(arr))
		for i, val := range arr {
			boolVal, err := ToBool(val)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element at index %d to bool: %w", i, err)
			}
			result[i] = boolVal
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
	}
}

// ToTypedArray is a generic function that converts a string array to an array of type T
// It handles common types (string, int, int64, float64, bool)
func ToTypedArray[T any](arr []string) ([]T, error) {
	if arr == nil {
		return nil, nil
	}

	result := make([]T, len(arr))

	// Get the type of T
	var zero T
	switch any(zero).(type) {
	case string:
		for i, val := range arr {
			result[i] = any(val).(T)
		}

	case int:
		for i, val := range arr {
			intVal, err := StringToInt(val)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element at index %d to int: %w", i, err)
			}
			result[i] = any(intVal).(T)
		}

	case int64:
		for i, val := range arr {
			int64Val, err := StringToInt64(val)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element at index %d to int64: %w", i, err)
			}
			result[i] = any(int64Val).(T)
		}

	case float64:
		for i, val := range arr {
			floatVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element at index %d to float64: %w", i, err)
			}
			result[i] = any(floatVal).(T)
		}

	case bool:
		for i, val := range arr {
			boolVal, err := ToBool(val)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element at index %d to bool: %w", i, err)
			}
			result[i] = any(boolVal).(T)
		}

	default:
		return nil, fmt.Errorf("unsupported type for conversion")
	}

	return result, nil
}

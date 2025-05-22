package logger

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"

	"github.com/ncobase/ncore/logging/logger/config"
	"github.com/sirupsen/logrus"
)

// Default patterns for detecting sensitive values
var defaultValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),          // Credit card
	regexp.MustCompile(`\b1[3-9]\d{9}\b`),                                     // Chinese mobile
	regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`), // Email
	regexp.MustCompile(`\b[A-Za-z0-9]{32,}\b`),                                // API keys/tokens
}

// Desensitizer handles sensitive data masking in log fields
type Desensitizer struct {
	config   *config.Desensitization
	patterns []*regexp.Regexp
}

// NewDesensitizer creates a new desensitizer instance
func NewDesensitizer(cfg *config.Desensitization) *Desensitizer {
	d := &Desensitizer{
		config:   cfg,
		patterns: make([]*regexp.Regexp, 0),
	}

	// Compile custom patterns
	for _, pattern := range cfg.CustomPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			d.patterns = append(d.patterns, regex)
		}
	}

	// Default patterns only if enabled
	if cfg.EnableDefaultPatterns {
		d.patterns = append(d.patterns, defaultValuePatterns...)
	}

	return d
}

// DesensitizeFields processes log fields and masks sensitive data
func (d *Desensitizer) DesensitizeFields(fields logrus.Fields) logrus.Fields {
	if !d.config.Enabled {
		return fields
	}

	result := make(logrus.Fields)
	for key, value := range fields {
		result[key] = d.desensitizeValue(key, value, 0)
	}
	return result
}

// desensitizeValue processes a single value recursively
func (d *Desensitizer) desensitizeValue(key string, value any, depth int) any {
	// Prevent infinite recursion
	if depth > 10 {
		return value
	}

	if value == nil {
		return nil
	}

	// Check field name sensitivity first
	if d.isSensitiveField(key) {
		return d.maskValue(value)
	}

	// Try reflection-based processing with panic recovery
	defer func() {
		if r := recover(); r != nil {
			// If reflection fails, return original value
			value = d.processViaJSON(value, depth)
		}
	}()

	// Process by type using reflection
	v := reflect.ValueOf(value)
	return d.processValueByType(key, v, depth)
}

// processViaJSON handles complex types via JSON marshaling (fallback method)
func (d *Desensitizer) processViaJSON(value any, depth int) any {
	// Convert to JSON and back to map[string]interface{} for safe processing
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return value // Return original if JSON marshal fails
	}

	var result any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return value // Return original if JSON unmarshal fails
	}

	// Process the result recursively
	return d.desensitizeValue("", result, depth+1)
}

// processValueByType handles different value types
func (d *Desensitizer) processValueByType(key string, v reflect.Value, depth int) any {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return d.processValueByType(key, v.Elem(), depth)
	}

	switch v.Kind() {
	case reflect.String:
		return d.desensitizeString(v.String())
	case reflect.Map:
		return d.processMap(v, depth)
	case reflect.Slice, reflect.Array:
		return d.processSlice(v, depth)
	case reflect.Struct:
		return d.processStruct(v, depth)
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return d.desensitizeValue(key, v.Interface(), depth+1)
	default:
		return v.Interface()
	}
}

// processMap handles map types recursively
func (d *Desensitizer) processMap(v reflect.Value, depth int) any {
	if v.IsNil() {
		return nil
	}

	newMap := reflect.MakeMap(v.Type())
	for _, key := range v.MapKeys() {
		mapValue := v.MapIndex(key)
		keyStr := d.reflectValueToString(key)
		processedValue := d.desensitizeValue(keyStr, mapValue.Interface(), depth+1)

		if processedValue != nil {
			processedReflectValue := reflect.ValueOf(processedValue)
			valueType := v.Type().Elem()

			// Handle type compatibility for map values
			if processedReflectValue.Type().AssignableTo(valueType) {
				newMap.SetMapIndex(key, processedReflectValue)
			} else if processedReflectValue.Type().ConvertibleTo(valueType) {
				newMap.SetMapIndex(key, processedReflectValue.Convert(valueType))
			} else {
				// If types don't match, keep original value
				newMap.SetMapIndex(key, mapValue)
			}
		} else {
			newMap.SetMapIndex(key, mapValue)
		}
	}
	return newMap.Interface()
}

// processSlice handles slice/array types recursively
func (d *Desensitizer) processSlice(v reflect.Value, depth int) any {
	length := v.Len()
	newSlice := reflect.MakeSlice(v.Type(), length, length)

	for i := 0; i < length; i++ {
		element := v.Index(i)
		processedElement := d.desensitizeValue("", element.Interface(), depth+1)

		// Handle type conversion safely
		if processedElement != nil {
			processedValue := reflect.ValueOf(processedElement)
			elementType := newSlice.Index(i).Type()

			// Check if types are compatible
			if processedValue.Type().AssignableTo(elementType) {
				newSlice.Index(i).Set(processedValue)
			} else if processedValue.Type().ConvertibleTo(elementType) {
				newSlice.Index(i).Set(processedValue.Convert(elementType))
			} else {
				// If types don't match, keep original value
				newSlice.Index(i).Set(element)
			}
		}
	}
	return newSlice.Interface()
}

// processStruct handles struct types recursively
func (d *Desensitizer) processStruct(v reflect.Value, depth int) any {
	// Try JSON marshaling for easier processing
	if jsonBytes, err := json.Marshal(v.Interface()); err == nil {
		var mapResult map[string]any
		if err := json.Unmarshal(jsonBytes, &mapResult); err == nil {
			return d.processMap(reflect.ValueOf(mapResult), depth)
		}
	}

	// Fallback: process struct fields directly
	structType := v.Type()
	newStruct := reflect.New(structType).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := structType.Field(i)

		if !field.CanInterface() {
			continue
		}

		// Get field name from JSON tag or struct field name
		fieldName := fieldType.Name
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
			if tagName := strings.Split(jsonTag, ",")[0]; tagName != "" && tagName != "-" {
				fieldName = tagName
			}
		}

		if field.CanSet() && newStruct.Field(i).CanSet() {
			processedValue := d.desensitizeValue(fieldName, field.Interface(), depth+1)
			if processedValue != nil {
				processedReflectValue := reflect.ValueOf(processedValue)
				fieldType := newStruct.Field(i).Type()

				// Check if types are compatible
				if processedReflectValue.Type().AssignableTo(fieldType) {
					newStruct.Field(i).Set(processedReflectValue)
				} else if processedReflectValue.Type().ConvertibleTo(fieldType) {
					newStruct.Field(i).Set(processedReflectValue.Convert(fieldType))
				} else {
					// If types don't match, keep original value
					newStruct.Field(i).Set(field)
				}
			} else {
				// Keep original field value if processed value is nil
				newStruct.Field(i).Set(field)
			}
		}
	}
	return newStruct.Interface()
}

// reflectValueToString converts reflect.Value to string for key comparison
func (d *Desensitizer) reflectValueToString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return string(rune(v.Int()))
	default:
		return ""
	}
}

// isSensitiveField checks if field name contains sensitive keywords
func (d *Desensitizer) isSensitiveField(fieldName string) bool {
	if fieldName == "" {
		return false
	}

	lowerName := strings.ToLower(fieldName)

	for _, sensitiveField := range d.config.SensitiveFields {
		lowerSensitiveField := strings.ToLower(sensitiveField)

		if d.config.ExactFieldMatch {
			// Exact match mode
			if lowerName == lowerSensitiveField {
				return true
			}
		} else {
			// Fuzzy match mode (contains)
			if strings.Contains(lowerName, lowerSensitiveField) {
				return true
			}
		}
	}
	return false
}

// desensitizeString applies pattern-based desensitization to strings
func (d *Desensitizer) desensitizeString(str string) string {
	if str == "" {
		return str
	}

	result := str
	for _, pattern := range d.patterns {
		if pattern.MatchString(result) {
			fixedMask := strings.Repeat(d.config.MaskChar, d.config.FixedMaskLength)
			result = pattern.ReplaceAllString(result, fixedMask)
		}
	}
	return result
}

// maskValue masks sensitive values with fixed-length replacement
func (d *Desensitizer) maskValue(value any) any {
	if value == nil {
		return nil
	}

	fixedMask := strings.Repeat(d.config.MaskChar, d.config.FixedMaskLength)

	switch v := value.(type) {
	case string:
		if v == "" {
			return v
		}
		return d.maskString(v)
	case []byte:
		if len(v) == 0 {
			return v
		}
		return fixedMask
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return fixedMask
	default:
		return fixedMask
	}
}

// maskString applies masking strategy based on configuration
func (d *Desensitizer) maskString(str string) string {
	if str == "" {
		return str
	}

	// Use fixed length masking for better security
	if d.config.UseFixedLength {
		return strings.Repeat(d.config.MaskChar, d.config.FixedMaskLength)
	}

	// Legacy mode: preserve prefix/suffix
	if d.config.PreservePrefix > 0 || d.config.PreserveSuffix > 0 {
		return d.maskStringWithPreserve(str)
	}

	// Default: fixed length
	return strings.Repeat(d.config.MaskChar, d.config.FixedMaskLength)
}

// maskStringWithPreserve preserves prefix/suffix characters (legacy mode)
func (d *Desensitizer) maskStringWithPreserve(str string) string {
	strLen := len(str)
	preserveTotal := d.config.PreservePrefix + d.config.PreserveSuffix

	if strLen <= preserveTotal {
		return strings.Repeat(d.config.MaskChar, d.config.FixedMaskLength)
	}

	prefix := str[:d.config.PreservePrefix]
	suffix := str[strLen-d.config.PreserveSuffix:]
	mask := strings.Repeat(d.config.MaskChar, d.config.FixedMaskLength)

	return prefix + mask + suffix
}

// DeepDesensitize provides standalone deep desensitization
func (d *Desensitizer) DeepDesensitize(data any) any {
	if !d.config.Enabled {
		return data
	}
	return d.desensitizeValue("", data, 0)
}

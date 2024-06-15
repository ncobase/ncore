package validator

import (
	"reflect"
)

// IsNil verify nil
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// IsNotNil verify is not nil
func IsNotNil(i any) bool {
	return !IsNil(i)
}

// IsEmpty verify is empty
func IsEmpty(i any) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		vi = vi.Elem()
	}
	switch vi.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return vi.Len() == 0
	case reflect.Bool:
		return !vi.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return vi.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return vi.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return vi.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return vi.IsNil()
	case reflect.Struct:
		return vi.NumField() == 0
	}
	return false
}

// IsNotEmpty verify is not empty
func IsNotEmpty(i any) bool {
	return !IsEmpty(i)
}

// RemoveEmptyString remove empty string of string array
func RemoveEmptyString(as []string) []string {
	var rs []string
	for _, s := range as {
		if !(IsEmpty(s) || IsNil(s)) {
			rs = append(rs, s)
		}
	}
	return rs
}

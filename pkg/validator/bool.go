package validator

import "reflect"

// BoolPtr convert bool to *bool
func BoolPtr(b bool) *bool {
	return &b
}

// PtrBool convert *bool to bool
func PtrBool(b *bool) bool {
	return *b
}

// IsTrue verify is true
func IsTrue(i any) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		vi = vi.Elem()
	}
	switch vi.Kind() {
	case reflect.Bool:
		return vi.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return vi.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return vi.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return vi.Float() != 0
	}
	return false
}

// IsFalse verify is false
func IsFalse(i any) bool {
	return !IsTrue(i)
}

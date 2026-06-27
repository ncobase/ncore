package validator

import "reflect"

// BoolPtr convert bool to *bool
func BoolPtr(b bool) *bool {
	return &b
}

// PtrBool convert *bool to bool
func PtrBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// IsTrue verify is true
func IsTrue(i any) bool {
	if i == nil {
		return false
	}
	vi := reflect.ValueOf(i)
	for vi.Kind() == reflect.Ptr || vi.Kind() == reflect.Interface {
		if vi.IsNil() {
			return false
		}
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

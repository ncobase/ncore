package validator

import "reflect"

// IsEqual verify is equal
func IsEqual(i, j any) bool {
	return reflect.DeepEqual(i, j)
}

// IsNotEqual verify is not equal
func IsNotEqual(i, j any) bool {
	return !IsEqual(i, j)
}

// IsGreater verify is greater
func IsGreater(i, j any) bool {
	ii := reflect.ValueOf(i)
	jj := reflect.ValueOf(j)
	if ii.Kind() == reflect.Ptr {
		ii = ii.Elem()
	}
	if jj.Kind() == reflect.Ptr {
		jj = jj.Elem()
	}
	switch ii.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ii.Int() > jj.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return ii.Uint() > jj.Uint()
	case reflect.Float32, reflect.Float64:
		return ii.Float() > jj.Float()
	}
	return false
}

// IsGreaterOrEqual verify is greater or equal
func IsGreaterOrEqual(i, j any) bool {
	return IsGreater(i, j) || IsEqual(i, j)
}

// IsLess verify is less
func IsLess(i, j any) bool {
	return !IsGreaterOrEqual(i, j)
}

// IsLessOrEqual verify is less or equal
func IsLessOrEqual(i, j any) bool {
	return !IsGreater(i, j)
}

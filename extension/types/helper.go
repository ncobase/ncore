package types

import "reflect"

// GetServiceInterface gets interface from service by field name
func GetServiceInterface(service any, fieldName string) (any, bool) {
	serviceValue := reflect.ValueOf(service)

	// Handle pointers
	if serviceValue.Kind() == reflect.Ptr {
		if serviceValue.IsNil() {
			return nil, false
		}
		serviceValue = serviceValue.Elem()
	}

	// Verify that the service is a struct
	if serviceValue.Kind() != reflect.Struct {
		return nil, false
	}

	// Try to find the field
	field := serviceValue.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, false
	}

	// Verify that the field is not nil
	if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
		if field.IsNil() {
			return nil, false
		}
	}

	return field.Interface(), true
}

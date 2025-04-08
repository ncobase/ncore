package types

// ToPointer is a helper function that returns a pointer to a value of any type.
func ToPointer[T any](v T) *T {
	return &v
}

// ToValue is a helper function that returns the value pointed to by a pointer.
func ToValue[T any](v *T) T {
	return *v
}

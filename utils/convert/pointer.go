package convert

// ToPointer returns a pointer to a value of any type
func ToPointer[T any](v T) *T {
	return &v
}

// ToValue returns the value pointed to by a pointer
func ToValue[T any](v *T) T {
	return *v
}

package validator

// IsIn verify is in
func IsIn(i any, j []any) bool {
	for _, v := range j {
		if IsEqual(i, v) {
			return true
		}
	}
	return false
}

// IsNotIn verify is not in
func IsNotIn(i any, j []any) bool {
	return !IsIn(i, j)
}

// IsContains verify is contains
func IsContains(i any, j []any) bool {
	for _, v := range j {
		if IsEqual(i, v) {
			return true
		}
	}
	return false
}

// IsNotContains verify is not contains
func IsNotContains(i any, j []any) bool {
	return !IsContains(i, j)
}

// IsInString verify is in string
func IsInString(i string, j []string) bool {
	for _, v := range j {
		if i == v {
			return true
		}
	}
	return false
}

// IsNotInString verify is not in string
func IsNotInString(i string, j []string) bool {
	return !IsInString(i, j)
}

// IsInArray verify is in array
func IsInArray(i any, j []any) bool {
	for _, v := range j {
		if IsEqual(i, v) {
			return true
		}
	}
	return false
}

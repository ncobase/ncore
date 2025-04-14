package ecode

import (
	"fmt"
)

const (
	emptyMsg       = "empty"
	requiredMsg    = "required"
	invalidMsg     = "invalid"
	successMsg     = "success"
	failedMsg      = "failed"
	existMsg       = "already exists"
	notExistMsg    = "does not exist"
	notSingularMsg = "not singular"
	expiredMsg     = "expired"
)

// FieldIsBlank returns field blank message
func FieldIsBlank(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], emptyMsg)
	}
	return emptyMsg
}

// FieldIsNil returns field nil message
func FieldIsNil(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], emptyMsg)
	}
	return emptyMsg
}

// FieldIsRequired returns field required message
func FieldIsRequired(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], requiredMsg)
	}
	return emptyMsg
}

// FieldIsEmpty returns field empty message
func FieldIsEmpty(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], emptyMsg)
	}
	return emptyMsg
}

// FieldIsInvalid returns field invalid message
func FieldIsInvalid(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], invalidMsg)
	}
	return invalidMsg
}

// Success returns success message
func Success(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], successMsg)
	}
	return successMsg
}

// Failed returns failed message
func Failed(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], failedMsg)
	}
	return failedMsg
}

// AlreadyExist returns already exist message
func AlreadyExist(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], existMsg)
	}
	return existMsg
}

// NotExist returns not exist message
func NotExist(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], notExistMsg)
	}
	return notExistMsg
}

// NotSingular returns not singular message
func NotSingular(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], notSingularMsg)
	}
	return notSingularMsg
}

// Expired returns expired message
func Expired(k ...string) string {
	if len(k) > 0 {
		return fmt.Sprintf("%s %s", k[0], expiredMsg)
	}
	return expiredMsg
}

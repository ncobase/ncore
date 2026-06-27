package validator

import "testing"

func TestNilHelpersDoNotPanic(t *testing.T) {
	if !IsEmpty(nil) {
		t.Fatal("nil should be empty")
	}
	if PtrBool(nil) {
		t.Fatal("nil bool pointer should be false")
	}
	if IsTrue(nil) {
		t.Fatal("nil should not be true")
	}
}

func TestIsEmptyDereferencesPointers(t *testing.T) {
	value := ""
	if !IsEmpty(&value) {
		t.Fatal("pointer to empty string should be empty")
	}

	value = "present"
	if IsEmpty(&value) {
		t.Fatal("pointer to non-empty string should not be empty")
	}
}

func TestValidateStructHandlesInvalidInputs(t *testing.T) {
	if errs := ValidateStruct(nil); errs["_"] == "" {
		t.Fatalf("expected nil input error, got %#v", errs)
	}
	if errs := ValidateStruct("not-struct"); errs["_"] == "" {
		t.Fatalf("expected invalid input error, got %#v", errs)
	}
}

func TestValidateStructSupportsStructValues(t *testing.T) {
	type request struct {
		Email string `json:"email" validate:"required,email"`
	}

	errs := ValidateStruct(request{})
	if errs["email"] == "" {
		t.Fatalf("expected email validation error, got %#v", errs)
	}
}

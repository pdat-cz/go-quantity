package quantity

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorIsMatchesByCode(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &Error{Code: CodeInexact, Op: "add"})
	if !errors.Is(err, &Error{Code: CodeInexact}) {
		t.Fatal("same code did not match")
	}
	if errors.Is(err, &Error{Code: CodeInvalidValue}) {
		t.Fatal("different code matched")
	}
	if !errors.Is(err, &Error{}) {
		t.Fatal("empty code should match any *Error")
	}
}

func TestErrorIsDoesNotPanicOnNil(t *testing.T) {
	err := &Error{Code: CodeInvalidValue}
	var typedNil *Error
	if errors.Is(err, typedNil) {
		t.Fatal("typed nil target matched")
	}
	if typedNil.Is(&Error{Code: CodeInvalidValue}) {
		t.Fatal("nil receiver matched")
	}
	if typedNil.Error() != "<nil>" || typedNil.Unwrap() != nil {
		t.Fatal("nil receiver methods changed behaviour")
	}
}

func TestErrorIsDoesNotUnwrapTarget(t *testing.T) {
	target := fmt.Errorf("outer: %w", &Error{Code: CodeInvalidValue})
	if errors.Is(&Error{Code: CodeInvalidValue}, target) {
		t.Fatal("Is inspected the target's wrapped chain")
	}
}

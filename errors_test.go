package quantity

import (
	"errors"
	"fmt"
	"strings"
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

func TestMalformedIdentifiersAreRejectedAndQuoted(t *testing.T) {
	forged := UnitID("length.metre\nforged=true")
	for name, call := range map[string]func() error{
		"New":        func() error { _, err := New("1", forged); return err },
		"NewFloat64": func() error { _, err := NewFloat64(1, forged); return err },
		"In": func() error {
			value, _ := New("1", UnitLengthMetre)
			_, err := value.In(forged)
			return err
		},
		"InFloat64": func() error {
			value, _ := New("1", UnitLengthMetre)
			_, err := value.InFloat64(forged)
			return err
		},
		"DataIn": func() error {
			value, _ := New("1", UnitLengthMetre)
			_, err := value.DataIn(forged)
			return err
		},
	} {
		err := call()
		if !errors.Is(err, &Error{Code: CodeInvalidID}) {
			t.Errorf("%s: error = %v, want invalid_id", name, err)
		}
		if strings.ContainsAny(err.Error(), "\n\r") {
			t.Errorf("%s: error text contains a raw newline: %q", name, err.Error())
		}
	}
}

func TestErrorQuotesIdentifiers(t *testing.T) {
	err := &Error{Code: CodeUnknownUnit, Op: "op", Unit: UnitID("a.b\n")}
	if got, want := err.Error(), `op: unknown_unit: unit "a.b\n"`; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	err = &Error{Code: CodeUnknownKind, Kind: Kind("k\x7f")}
	if got, want := err.Error(), `unknown_kind: kind "k\x7f"`; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

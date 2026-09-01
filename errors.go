package quantity

import (
	"errors"
	"fmt"
)

// Code is a stable, machine-readable error category.
type Code string

const (
	CodeInvalidID         Code = "invalid_id"
	CodeInvalidDefinition Code = "invalid_definition"
	CodeDuplicateKind     Code = "duplicate_kind"
	CodeDuplicateUnit     Code = "duplicate_unit"
	CodeUnknownKind       Code = "unknown_kind"
	CodeUnknownUnit       Code = "unknown_unit"
	CodeKindMismatch      Code = "kind_mismatch"
	CodeInvalidValue      Code = "invalid_value"
	CodeInexact           Code = "inexact"
)

// Error is returned for model, catalog, and conversion failures.
type Error struct {
	Code Code
	Op   string
	Kind Kind
	Unit UnitID
	Err  error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	detail := string(e.Code)
	if e.Unit != "" {
		detail += ": unit " + e.Unit.String()
	} else if e.Kind != "" {
		detail += ": kind " + e.Kind.String()
	}
	if e.Err != nil {
		detail += ": " + e.Err.Error()
	}
	if e.Op == "" {
		return detail
	}
	return fmt.Sprintf("%s: %s", e.Op, detail)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is lets callers match errors by Code with errors.Is.
func (e *Error) Is(target error) bool {
	var other *Error
	return errors.As(target, &other) && (other.Code == "" || e.Code == other.Code)
}

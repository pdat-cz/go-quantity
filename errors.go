package quantity

import (
	"fmt"
	"strconv"
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
		detail += ": unit " + strconv.Quote(e.Unit.String())
	} else if e.Kind != "" {
		detail += ": kind " + strconv.Quote(e.Kind.String())
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

// Is lets callers match errors by Code with errors.Is. A target with an empty
// Code matches any *Error. Only the target itself is inspected, never its
// wrapped chain, as the errors.Is contract requires.
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	if !ok || other == nil || e == nil {
		return false
	}
	return other.Code == "" || e.Code == other.Code
}

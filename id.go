package quantity

import (
	"errors"
	"strings"
)

const (
	maxKindIDBytes = 128
	maxUnitIDBytes = 192
)

// Kind identifies the semantic kind of a quantity, for example
// electric_current or temperature_difference.
type Kind string

// ParseKind validates and returns an OQS QuantityKind identifier.
func ParseKind(text string) (Kind, error) {
	if !validIDSegment(text, maxKindIDBytes) {
		return "", &Error{Code: CodeInvalidID, Op: "parse kind", Err: errors.New("expected lowercase ASCII identifier")}
	}
	return Kind(text), nil
}

func (k Kind) String() string { return string(k) }

// Valid reports whether k is a valid OQS QuantityKind identifier.
func (k Kind) Valid() bool {
	_, err := ParseKind(k.String())
	return err == nil
}

// UnitID is a stable identifier scoped by Kind, for example
// electric_current.ampere.
type UnitID string

// ParseUnitID validates and returns an OQS UnitID.
func ParseUnitID(text string) (UnitID, error) {
	if len(text) == 0 || len(text) > maxUnitIDBytes || strings.Count(text, ".") != 1 {
		return "", &Error{Code: CodeInvalidID, Op: "parse unit id", Err: errors.New("expected quantity_kind.unit")}
	}
	kind, unit, _ := strings.Cut(text, ".")
	if !validIDSegment(kind, maxKindIDBytes) || !validIDSegment(unit, maxKindIDBytes) {
		return "", &Error{Code: CodeInvalidID, Op: "parse unit id", Err: errors.New("expected lowercase ASCII identifier segments")}
	}
	return UnitID(text), nil
}

func (id UnitID) String() string { return string(id) }

// Valid reports whether id is a valid OQS UnitID.
func (id UnitID) Valid() bool {
	_, err := ParseUnitID(id.String())
	return err == nil
}

// Kind returns the QuantityKind encoded in id.
func (id UnitID) Kind() (Kind, error) {
	parsed, err := ParseUnitID(id.String())
	if err != nil {
		return "", err
	}
	kind, _, _ := strings.Cut(parsed.String(), ".")
	return Kind(kind), nil
}

func validIDSegment(text string, max int) bool {
	if len(text) == 0 || len(text) > max || text[0] < 'a' || text[0] > 'z' {
		return false
	}
	for i := 1; i < len(text); i++ {
		char := text[i]
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '_' {
			return false
		}
	}
	return true
}

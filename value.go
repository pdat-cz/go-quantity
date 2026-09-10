package quantity

import (
	"encoding/json/v2"
	"errors"
	"math"
)

const maxValueJSONBytes = 4096

// Value is the Go representation of an OQS QuantityValue. Its amount and Unit
// are always stored in the QuantityKind's ReferenceUnit.
type Value struct {
	amount  Decimal
	unit    UnitID
	kind    Kind
	catalog *Catalog
}

// New parses amount and creates a Value using StandardCatalog.
func New(amount string, unit UnitID) (Value, error) {
	decimal, err := ParseDecimal(amount)
	if err != nil {
		return Value{}, err
	}
	return StandardCatalog.New(decimal, unit)
}

// NewFloat64 creates a Value at the explicit binary-float approximation
// boundary using StandardCatalog.
func NewFloat64(amount float64, unit UnitID) (Value, error) {
	return StandardCatalog.NewFloat64(amount, unit)
}

// New converts amount from unit to the appropriate ReferenceUnit.
func (c *Catalog) New(amount Decimal, unit UnitID) (Value, error) {
	definition, kind, err := c.input(unit, "new value")
	if err != nil {
		return Value{}, err
	}
	reference, err := definition.toReference.toReferenceExact(amount)
	if err != nil {
		return Value{}, wrapConversionError("new value", kind.id, unit, err)
	}
	return Value{amount: reference, unit: kind.referenceUnit, kind: kind.id, catalog: c}, nil
}

// NewFloat64 converts a finite binary float from unit to ReferenceUnit. The
// result is stored as the shortest round-trippable decimal representation.
func (c *Catalog) NewFloat64(amount float64, unit UnitID) (Value, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return Value{}, valueError("new float64 value", "value must be finite")
	}
	definition, kind, err := c.input(unit, "new float64 value")
	if err != nil {
		return Value{}, err
	}
	referenceFloat, err := definition.toReference.toReferenceFloat64(amount)
	if err != nil || math.IsNaN(referenceFloat) || math.IsInf(referenceFloat, 0) {
		return Value{}, wrapConversionError("new float64 value", kind.id, unit, err)
	}
	reference, err := DecimalFromFloat64(referenceFloat)
	if err != nil {
		return Value{}, err
	}
	return Value{amount: reference, unit: kind.referenceUnit, kind: kind.id, catalog: c}, nil
}

// Decimal returns the exact amount stored in ReferenceUnit.
func (v Value) Decimal() Decimal { return v.amount }

// Unit returns the ReferenceUnit in which Decimal is stored.
func (v Value) Unit() UnitID { return v.unit }

// Kind returns the value's semantic QuantityKind.
func (v Value) Kind() Kind { return v.kind }

// Data returns the minimal transport representation in ReferenceUnit.
func (v Value) Data() Data {
	return Data{Value: v.amount.String(), Unit: v.unit}
}

// DataIn returns the minimal transport representation expressed in unit.
func (v Value) DataIn(unit UnitID) (Data, error) {
	amount, err := v.In(unit)
	if err != nil {
		return Data{}, err
	}
	return Data{Value: amount.String(), Unit: unit}, nil
}

// In returns v expressed exactly in unit. It fails if the result has no finite
// decimal representation.
func (v Value) In(unit UnitID) (Decimal, error) {
	if err := v.valid("convert value"); err != nil {
		return Decimal{}, err
	}
	target, err := v.catalog.lookup(unit, "convert value")
	if err != nil {
		return Decimal{}, err
	}
	if target.kind != v.kind {
		return Decimal{}, &Error{Code: CodeKindMismatch, Op: "convert value", Kind: v.kind, Unit: unit}
	}
	converted, err := target.toReference.fromReferenceExact(v.amount)
	if err != nil {
		return Decimal{}, wrapConversionError("convert value", v.kind, unit, err)
	}
	return converted, nil
}

// InFloat64 returns v expressed in unit at the explicit binary-float
// approximation boundary.
func (v Value) InFloat64(unit UnitID) (float64, error) {
	if err := v.valid("convert value to float64"); err != nil {
		return 0, err
	}
	target, err := v.catalog.lookup(unit, "convert value to float64")
	if err != nil {
		return 0, err
	}
	if target.kind != v.kind {
		return 0, &Error{Code: CodeKindMismatch, Op: "convert value to float64", Kind: v.kind, Unit: unit}
	}
	reference, _ := v.amount.Float64()
	converted, err := target.toReference.fromReferenceFloat64(reference)
	if err != nil || math.IsNaN(converted) || math.IsInf(converted, 0) {
		return 0, wrapConversionError("convert value to float64", v.kind, unit, err)
	}
	return converted, nil
}

// MarshalJSON writes the minimal OQS JSON form.
func (v Value) MarshalJSON() ([]byte, error) {
	if err := v.valid("marshal value"); err != nil {
		return nil, err
	}
	return json.Marshal(v.Data())
}

// UnmarshalJSON reads the minimal OQS JSON form using StandardCatalog.
func (v *Value) UnmarshalJSON(data []byte) error {
	parsed, err := StandardCatalog.ParseJSON(data)
	if err != nil {
		return err
	}
	*v = parsed
	return nil
}

// ParseJSON reads the minimal OQS JSON form with this catalog. Whitespace and
// object field order are insignificant; unknown fields are rejected.
func (c *Catalog) ParseJSON(data []byte) (Value, error) {
	if len(data) == 0 || len(data) > maxValueJSONBytes {
		return Value{}, valueError("parse value JSON", "invalid input length")
	}
	var decoded Data
	if err := json.Unmarshal(data, &decoded); err != nil {
		return Value{}, &Error{Code: CodeInvalidValue, Op: "parse value JSON", Err: err}
	}
	return c.Decode(decoded)
}

func (c *Catalog) input(unit UnitID, op string) (UnitDefinition, KindDefinition, error) {
	if c == nil {
		return UnitDefinition{}, KindDefinition{}, &Error{Code: CodeInvalidDefinition, Op: op, Err: errors.New("nil catalog")}
	}
	definition, err := c.lookup(unit, op)
	if err != nil {
		return UnitDefinition{}, KindDefinition{}, err
	}
	kind, ok := c.Kind(definition.kind)
	if !ok {
		return UnitDefinition{}, KindDefinition{}, &Error{Code: CodeUnknownKind, Op: op, Kind: definition.kind, Unit: unit}
	}
	return definition, kind, nil
}

// lookup validates unit syntactically before touching the catalog so that a
// malformed identifier is reported as CodeInvalidID and never echoed verbatim.
func (c *Catalog) lookup(unit UnitID, op string) (UnitDefinition, error) {
	if _, err := ParseUnitID(unit.String()); err != nil {
		return UnitDefinition{}, &Error{Code: CodeInvalidID, Op: op, Err: errors.Unwrap(err)}
	}
	definition, ok := c.Unit(unit)
	if !ok {
		return UnitDefinition{}, &Error{Code: CodeUnknownUnit, Op: op, Unit: unit}
	}
	return definition, nil
}

func (v Value) valid(op string) error {
	if v.catalog == nil || !v.unit.Valid() || !v.kind.Valid() {
		return &Error{Code: CodeInvalidValue, Op: op, Unit: v.unit, Err: errors.New("zero or invalid Value")}
	}
	kind, ok := v.catalog.Kind(v.kind)
	if !ok || kind.referenceUnit != v.unit {
		return &Error{Code: CodeInvalidValue, Op: op, Kind: v.kind, Unit: v.unit, Err: errors.New("Value is not stored in ReferenceUnit")}
	}
	return nil
}

func wrapConversionError(op string, kind Kind, unit UnitID, err error) error {
	if err == nil {
		err = errors.New("conversion produced a non-finite value")
	}
	var typed *Error
	if errors.As(err, &typed) && typed.Code == CodeInexact {
		return &Error{Code: CodeInexact, Op: op, Kind: kind, Unit: unit, Err: typed.Err}
	}
	return &Error{Code: CodeInvalidValue, Op: op, Kind: kind, Unit: unit, Err: err}
}

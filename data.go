package quantity

import (
	"encoding/json/v2"
	"errors"
)

// Data is the representation-neutral transport form of an OQS QuantityValue.
// Value is a string so encoders cannot silently lose decimal precision.
//
// The yaml tags are metadata only and do not add a YAML dependency.
type Data struct {
	Value string `json:"value,case:strict" yaml:"value"`
	Unit  UnitID `json:"unit,case:strict" yaml:"unit"`
}

// Validate checks the representation without resolving Unit against a Catalog.
func (d Data) Validate() error {
	if _, err := ParseDecimal(d.Value); err != nil {
		return err
	}
	if _, err := ParseUnitID(d.Unit.String()); err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON uses Go 1.27 JSON v2 semantics. It rejects unknown members,
// duplicate names, invalid UTF-8, and case variants of the defined names.
func (d *Data) UnmarshalJSON(input []byte) error {
	if d == nil {
		return &Error{Code: CodeInvalidValue, Op: "unmarshal data", Err: errors.New("nil Data receiver")}
	}
	if len(input) == 0 || len(input) > maxValueJSONBytes {
		return valueError("unmarshal data", "invalid input length")
	}
	type plainData Data
	var decoded plainData
	if err := json.Unmarshal(input, &decoded, json.RejectUnknownMembers(true)); err != nil {
		return &Error{Code: CodeInvalidValue, Op: "unmarshal data", Err: err}
	}
	if err := Data(decoded).Validate(); err != nil {
		return err
	}
	*d = Data(decoded)
	return nil
}

// Decode validates data against c and converts it to ReferenceUnit.
func (c *Catalog) Decode(data Data) (Value, error) {
	if err := data.Validate(); err != nil {
		return Value{}, err
	}
	amount, err := ParseDecimal(data.Value)
	if err != nil {
		return Value{}, err
	}
	return c.New(amount, data.Unit)
}

package quantity

import (
	"errors"
	"math/big"
)

// Transform is an exact affine conversion to a Kind's ReferenceUnit:
// reference = value*scale + offset.
type Transform struct {
	scale  *big.Rat
	offset *big.Rat
}

// NewTransform constructs an exact affine transform from two rational
// numbers. A scale of zero is invalid.
func NewTransform(scaleNumerator, scaleDenominator, offsetNumerator, offsetDenominator string) (Transform, error) {
	scale, err := parseRational(scaleNumerator, scaleDenominator)
	if err != nil || scale.Sign() == 0 {
		return Transform{}, &Error{Code: CodeInvalidDefinition, Op: "new transform", Err: errors.New("scale must be a non-zero rational number")}
	}
	offset, err := parseRational(offsetNumerator, offsetDenominator)
	if err != nil {
		return Transform{}, &Error{Code: CodeInvalidDefinition, Op: "new transform", Err: errors.New("offset must be a rational number")}
	}
	return Transform{scale: scale, offset: offset}, nil
}

// IdentityTransform returns a transform that leaves a value unchanged.
func IdentityTransform() Transform {
	return Transform{scale: big.NewRat(1, 1), offset: new(big.Rat)}
}

// Scale returns the normalized numerator and denominator of the scale.
func (t Transform) Scale() (string, string) {
	if t.scale == nil {
		return "0", "1"
	}
	return t.scale.Num().String(), t.scale.Denom().String()
}

// Offset returns the normalized numerator and denominator of the offset.
func (t Transform) Offset() (string, string) {
	if t.offset == nil {
		return "0", "1"
	}
	return t.offset.Num().String(), t.offset.Denom().String()
}

// IsIdentity reports whether t leaves a value unchanged.
func (t Transform) IsIdentity() bool {
	return t.valid() && t.scale.Cmp(big.NewRat(1, 1)) == 0 && t.offset.Sign() == 0
}

func (t Transform) valid() bool {
	return t.scale != nil && t.offset != nil && t.scale.Sign() != 0
}

func (t Transform) toReferenceExact(value Decimal) (Decimal, error) {
	if !t.valid() {
		return Decimal{}, &Error{Code: CodeInvalidDefinition, Op: "convert to reference", Err: errors.New("invalid transform")}
	}
	result := new(big.Rat).Mul(value.rat(), t.scale)
	result.Add(result, t.offset)
	return decimalFromRatExact(result, "convert to reference")
}

func (t Transform) fromReferenceExact(value Decimal) (Decimal, error) {
	if !t.valid() {
		return Decimal{}, &Error{Code: CodeInvalidDefinition, Op: "convert from reference", Err: errors.New("invalid transform")}
	}
	result := new(big.Rat).Sub(value.rat(), t.offset)
	result.Quo(result, t.scale)
	return decimalFromRatExact(result, "convert from reference")
}

func (t Transform) toReferenceFloat64(value float64) (float64, error) {
	if !t.valid() {
		return 0, &Error{Code: CodeInvalidDefinition, Op: "convert to reference float64", Err: errors.New("invalid transform")}
	}
	scale, _ := t.scale.Float64()
	offset, _ := t.offset.Float64()
	return value*scale + offset, nil
}

func (t Transform) fromReferenceFloat64(value float64) (float64, error) {
	if !t.valid() {
		return 0, &Error{Code: CodeInvalidDefinition, Op: "convert from reference float64", Err: errors.New("invalid transform")}
	}
	scale, _ := t.scale.Float64()
	offset, _ := t.offset.Float64()
	return (value - offset) / scale, nil
}

func parseRational(numerator, denominator string) (*big.Rat, error) {
	n, ok := new(big.Int).SetString(numerator, 10)
	if !ok {
		return nil, errors.New("invalid numerator")
	}
	d, ok := new(big.Int).SetString(denominator, 10)
	if !ok || d.Sign() == 0 {
		return nil, errors.New("invalid denominator")
	}
	if d.Sign() < 0 {
		n.Neg(n)
		d.Neg(d)
	}
	return new(big.Rat).SetFrac(n, d), nil
}

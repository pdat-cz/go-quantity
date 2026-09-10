package quantity

import (
	"errors"
	"math"
	"math/big"
	"strconv"
	"strings"
)

const (
	maxDecimalDigits     = 100_000
	maxExponentMagnitude = 100_000
)

// Decimal is an immutable finite decimal represented as coefficient ×
// 10^exponent. Its zero value is the number zero.
type Decimal struct {
	coefficient *big.Int
	exponent    int32
}

// NewDecimal constructs and normalizes coefficient × 10^exponent.
func NewDecimal(coefficient string, exponent int32) (Decimal, error) {
	if exponent < -maxExponentMagnitude || exponent > maxExponentMagnitude {
		return Decimal{}, valueError("new decimal", "exponent is out of range")
	}
	digits := strings.TrimPrefix(coefficient, "-")
	if len(digits) > maxDecimalDigits {
		return Decimal{}, valueError("new decimal", "coefficient is too long")
	}
	if !isDigits(digits) {
		return Decimal{}, valueError("new decimal", "invalid coefficient")
	}
	integer, _ := new(big.Int).SetString(coefficient, 10)
	return normalizeDecimal(integer, int64(exponent), "new decimal")
}

// ParseDecimal parses a finite base-10 number. Plain and scientific notation
// are accepted and normalized.
func ParseDecimal(text string) (Decimal, error) {
	if text == "" || len(text) > maxDecimalDigits+32 {
		return Decimal{}, valueError("parse decimal", "invalid length")
	}
	i := 0
	negative := false
	if text[i] == '-' || text[i] == '+' {
		negative = text[i] == '-'
		i++
	}
	if i == len(text) {
		return Decimal{}, valueError("parse decimal", "missing digits")
	}
	integerStart := i
	for i < len(text) && text[i] >= '0' && text[i] <= '9' {
		i++
	}
	if i == integerStart {
		return Decimal{}, valueError("parse decimal", "missing integer digits")
	}
	digits := text[integerStart:i]
	fractionDigits := 0
	if i < len(text) && text[i] == '.' {
		i++
		fractionStart := i
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			i++
		}
		if i == fractionStart {
			return Decimal{}, valueError("parse decimal", "missing fractional digits")
		}
		fractionDigits = i - fractionStart
		digits += text[fractionStart:i]
	}
	exponent := int64(-fractionDigits)
	if i < len(text) && (text[i] == 'e' || text[i] == 'E') {
		i++
		exponentNegative := false
		if i < len(text) && (text[i] == '-' || text[i] == '+') {
			exponentNegative = text[i] == '-'
			i++
		}
		exponentStart := i
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			i++
		}
		if i == exponentStart || i-exponentStart > 7 {
			return Decimal{}, valueError("parse decimal", "invalid exponent")
		}
		parsed, err := strconv.ParseInt(text[exponentStart:i], 10, 32)
		if err != nil {
			return Decimal{}, valueError("parse decimal", "invalid exponent")
		}
		if exponentNegative {
			parsed = -parsed
		}
		exponent += parsed
	}
	if i != len(text) {
		return Decimal{}, valueError("parse decimal", "unexpected character")
	}
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return Decimal{}, nil
	}
	if len(digits) > maxDecimalDigits {
		return Decimal{}, valueError("parse decimal", "coefficient is too long")
	}
	if negative {
		digits = "-" + digits
	}
	coefficient, _ := new(big.Int).SetString(digits, 10)
	return normalizeDecimal(coefficient, exponent, "parse decimal")
}

// DecimalFromFloat64 converts the shortest round-trippable representation of
// value to Decimal. It is an explicit approximation boundary.
func DecimalFromFloat64(value float64) (Decimal, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return Decimal{}, valueError("decimal from float64", "value must be finite")
	}
	return ParseDecimal(strconv.FormatFloat(value, 'g', -1, 64))
}

// Coefficient returns the canonical base-10 coefficient.
func (d Decimal) Coefficient() string {
	if d.coefficient == nil {
		return "0"
	}
	return d.coefficient.String()
}

// Exponent returns the power of ten applied to Coefficient.
func (d Decimal) Exponent() int32 { return d.exponent }

// String returns the exact value without exponent notation.
func (d Decimal) String() string {
	if d.coefficient == nil || d.coefficient.Sign() == 0 {
		return "0"
	}
	negative := d.coefficient.Sign() < 0
	digits := new(big.Int).Abs(d.coefficient).String()
	exponent := int(d.exponent)
	var text string
	switch {
	case exponent >= 0:
		text = digits + strings.Repeat("0", exponent)
	case len(digits)+exponent > 0:
		point := len(digits) + exponent
		text = digits[:point] + "." + digits[point:]
	default:
		text = "0." + strings.Repeat("0", -exponent-len(digits)) + digits
	}
	if negative {
		return "-" + text
	}
	return text
}

// Float64 converts d to a binary float and reports whether the conversion is
// exact.
func (d Decimal) Float64() (float64, bool) { return d.rat().Float64() }

// Add returns the exact sum.
func (d Decimal) Add(other Decimal) (Decimal, error) {
	return d.addScaled(other, 1, "add")
}

// Sub returns the exact difference.
func (d Decimal) Sub(other Decimal) (Decimal, error) {
	return d.addScaled(other, -1, "subtract")
}

// addScaled computes d + sign*other by aligning coefficients to the smaller
// exponent. It never goes through big.Rat, so no factor stripping is needed.
func (d Decimal) addScaled(other Decimal, sign int64, op string) (Decimal, error) {
	if other.coefficient == nil || other.coefficient.Sign() == 0 {
		return d, nil
	}
	if d.coefficient == nil || d.coefficient.Sign() == 0 {
		return normalizeDecimal(new(big.Int).Mul(other.coefficient, big.NewInt(sign)), int64(other.exponent), op)
	}
	left := new(big.Int).Set(d.coefficient)
	right := new(big.Int).Mul(other.coefficient, big.NewInt(sign))
	exponent := int64(d.exponent)
	switch {
	case d.exponent > other.exponent:
		left.Mul(left, pow10(int(d.exponent-other.exponent)))
		exponent = int64(other.exponent)
	case d.exponent < other.exponent:
		right.Mul(right, pow10(int(other.exponent-d.exponent)))
	}
	return normalizeDecimal(left.Add(left, right), exponent, op)
}

// Cmp compares d and other.
func (d Decimal) Cmp(other Decimal) int { return d.rat().Cmp(other.rat()) }

func (d Decimal) rat() *big.Rat {
	if d.coefficient == nil {
		return new(big.Rat)
	}
	numerator := new(big.Int).Set(d.coefficient)
	if d.exponent >= 0 {
		numerator.Mul(numerator, pow10(int(d.exponent)))
		return new(big.Rat).SetInt(numerator)
	}
	return new(big.Rat).SetFrac(numerator, pow10(-int(d.exponent)))
}

func decimalFromRatExact(value *big.Rat, op string) (Decimal, error) {
	if value == nil || value.Sign() == 0 {
		return Decimal{}, nil
	}
	numerator := new(big.Int).Set(value.Num())
	denominator := new(big.Int).Set(value.Denom())
	twos := int(denominator.TrailingZeroBits())
	denominator.Rsh(denominator, uint(twos))
	fives := stripFactor(denominator, 5)
	if denominator.Cmp(bigOne) != 0 {
		return Decimal{}, &Error{Code: CodeInexact, Op: op, Err: errors.New("result has no finite decimal representation")}
	}
	scale := max(twos, fives)
	if scale > maxExponentMagnitude {
		return Decimal{}, valueError(op, "result exponent is out of range")
	}
	if twos < scale {
		numerator.Mul(numerator, new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(scale-twos)), nil))
	}
	if fives < scale {
		numerator.Mul(numerator, new(big.Int).Exp(big.NewInt(5), big.NewInt(int64(scale-fives)), nil))
	}
	return normalizeDecimal(numerator, int64(-scale), op)
}

var bigOne = big.NewInt(1)

// stripFactor divides n by factor as many times as it divides evenly and
// returns that count. It removes factors in chunks that fit a uint64 so that
// a denominator like 5^100000 costs thousands, not hundreds of thousands, of
// big-integer divisions.
func stripFactor(n *big.Int, factor int64) int {
	const chunk = 27 // 5^27 < 2^63
	chunkValue := new(big.Int).Exp(big.NewInt(factor), big.NewInt(chunk), nil)
	single := big.NewInt(factor)
	count := 0
	quotient, remainder := new(big.Int), new(big.Int)
	for {
		quotient.QuoRem(n, chunkValue, remainder)
		if remainder.Sign() != 0 {
			break
		}
		n.Set(quotient)
		count += chunk
	}
	for {
		quotient.QuoRem(n, single, remainder)
		if remainder.Sign() != 0 {
			break
		}
		n.Set(quotient)
		count++
	}
	return count
}

func normalizeDecimal(coefficient *big.Int, exponent int64, op string) (Decimal, error) {
	if coefficient == nil || coefficient.Sign() == 0 {
		return Decimal{}, nil
	}
	digits := new(big.Int).Abs(coefficient).String()
	trailing := len(digits) - len(strings.TrimRight(digits, "0"))
	if trailing > 0 {
		coefficient = new(big.Int).Quo(coefficient, pow10(trailing))
		exponent += int64(trailing)
	}
	if exponent < -maxExponentMagnitude || exponent > maxExponentMagnitude {
		return Decimal{}, valueError(op, "exponent is out of range")
	}
	if len(digits)-trailing > maxDecimalDigits {
		return Decimal{}, valueError(op, "coefficient is too long")
	}
	return Decimal{coefficient: new(big.Int).Set(coefficient), exponent: int32(exponent)}, nil
}

// isDigits reports whether text is a non-empty run of ASCII digits. It is
// checked before big.Int.SetString so that length limits apply before any
// allocation proportional to the input.
func isDigits(text string) bool {
	if text == "" {
		return false
	}
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

func pow10(exponent int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
}

func valueError(op, message string) error {
	return &Error{Code: CodeInvalidValue, Op: op, Err: errors.New(message)}
}

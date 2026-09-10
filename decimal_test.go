package quantity

import (
	jsonv1 "encoding/json"
	json "encoding/json/v2"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseDecimalNormalizes(t *testing.T) {
	value, err := ParseDecimal("001.2300e2")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := value.String(), "123"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if got, want := value.Coefficient(), "123"; got != want {
		t.Fatalf("Coefficient() = %q, want %q", got, want)
	}
	if value.Exponent() != 0 {
		t.Fatalf("Exponent() = %d, want 0", value.Exponent())
	}
}

func TestDecimalExactArithmetic(t *testing.T) {
	left, err := ParseDecimal("1.625")
	if err != nil {
		t.Fatal(err)
	}
	right, err := ParseDecimal("0.375")
	if err != nil {
		t.Fatal(err)
	}
	sum, err := left.Add(right)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := sum.String(), "2"; got != want {
		t.Fatalf("sum = %q, want %q", got, want)
	}
}

func TestDecimalSubAndSignedArithmetic(t *testing.T) {
	cases := []struct{ left, right, sum, difference string }{
		{"1.625", "0.375", "2", "1.25"},
		{"-1.5", "1.5", "0", "-3"},
		{"0", "-2.5", "-2.5", "2.5"},
		{"1e10", "1e-10", "10000000000.0000000001", "9999999999.9999999999"},
		{"123", "0", "123", "123"},
	}
	for _, c := range cases {
		left, err := ParseDecimal(c.left)
		if err != nil {
			t.Fatal(err)
		}
		right, err := ParseDecimal(c.right)
		if err != nil {
			t.Fatal(err)
		}
		sum, err := left.Add(right)
		if err != nil {
			t.Fatal(err)
		}
		if sum.String() != c.sum {
			t.Errorf("%s + %s = %s, want %s", c.left, c.right, sum, c.sum)
		}
		difference, err := left.Sub(right)
		if err != nil {
			t.Fatal(err)
		}
		if difference.String() != c.difference {
			t.Errorf("%s - %s = %s, want %s", c.left, c.right, difference, c.difference)
		}
	}
}

// TestExtremeExponentsAreCheap guards against the O(n) big-integer division
// loops that made a 45-byte JSON payload cost seconds of CPU.
func TestExtremeExponentsAreCheap(t *testing.T) {
	const budget = 500 * time.Millisecond
	cases := []struct {
		text string
		unit UnitID
	}{
		{"1e-1500", UnitLengthKilometre},
		{"1e1500", UnitLengthMillimetre},
		{"-7e-1499", UnitLengthKilometre},
	}
	for _, c := range cases {
		text := c.text
		start := time.Now()
		value, err := ParseDecimal(text)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := value.Add(Decimal{}); err != nil {
			t.Fatal(err)
		}
		if _, err := value.Sub(value); err != nil {
			t.Fatal(err)
		}
		if _, err := StandardCatalog.New(value, UnitLengthMetre); err != nil {
			t.Fatal(err)
		}
		if _, err := StandardCatalog.New(value, c.unit); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed > budget {
			t.Fatalf("%s took %s, budget %s", text, elapsed, budget)
		}
	}
}

func TestNewDecimalRejectsBadCoefficientBeforeParsing(t *testing.T) {
	cases := map[string]string{
		"empty":        "",
		"plus sign":    "+1",
		"double minus": "--1",
		"hex":          "0x10",
		"underscore":   "1_000",
		"space":        " 1",
		"too long":     strings.Repeat("9", maxDecimalDigits+1),
		"huge":         strings.Repeat("9", 1_000_000),
	}
	for name, coefficient := range cases {
		start := time.Now()
		if _, err := NewDecimal(coefficient, 0); !errors.Is(err, &Error{Code: CodeInvalidValue}) {
			t.Errorf("%s: NewDecimal(%q) error = %v, want invalid_value", name, coefficient, err)
		}
		if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
			t.Errorf("%s: rejection took %s", name, elapsed)
		}
	}
	for _, coefficient := range []string{"0", "-0", "1", "-1", "007", strings.Repeat("9", maxDecimalDigits)} {
		if _, err := NewDecimal(coefficient, 0); err != nil {
			t.Errorf("NewDecimal(%q) = %v, want ok", coefficient, err)
		}
	}
}

func TestDecimalStringRoundTripsAtLimits(t *testing.T) {
	nines := strings.Repeat("9", maxDecimalDigits)
	cases := []string{
		"1e1500", "1e-1500", "-1e1500",
		nines + "e1500", nines + "e-1500", "-" + nines + "e-1500",
		"1" + strings.Repeat("0", 1500), // trailing zeros do not count as coefficient digits
		"0." + strings.Repeat("0", 1499) + "9",
	}
	for _, text := range cases {
		value, err := ParseDecimal(text)
		if err != nil {
			t.Fatalf("ParseDecimal(%.20s…): %v", text, err)
		}
		printed := value.String()
		if len(printed) > maxDecimalTextBytes {
			t.Fatalf("String() of %.20s… is %d bytes, over the parse limit", text, len(printed))
		}
		again, err := ParseDecimal(printed)
		if err != nil {
			t.Fatalf("ParseDecimal(String()) of %.20s…: %v", text, err)
		}
		if again.Cmp(value) != 0 {
			t.Fatalf("round trip of %.20s… changed the value", text)
		}
	}
	for _, text := range []string{"1e1501", "1e-1501", nines + "9", "1" + nines + "e-1"} {
		if _, err := ParseDecimal(text); !errors.Is(err, &Error{Code: CodeInvalidValue}) {
			t.Errorf("ParseDecimal(%.20s…) error = %v, want invalid_value", text, err)
		}
	}
}

func TestDecimalSerializesAsTextInBothJSONVersions(t *testing.T) {
	type reading struct {
		Amount Decimal `json:"amount"`
	}
	amount, err := ParseDecimal("12.340")
	if err != nil {
		t.Fatal(err)
	}
	want := `{"amount":"12.34"}`

	v1, err := jsonv1.Marshal(reading{Amount: amount})
	if err != nil {
		t.Fatal(err)
	}
	if string(v1) != want {
		t.Fatalf("encoding/json v1 = %s, want %s", v1, want)
	}
	var decodedV1 reading
	if err := jsonv1.Unmarshal(v1, &decodedV1); err != nil {
		t.Fatal(err)
	}
	if decodedV1.Amount.Cmp(amount) != 0 {
		t.Fatal("encoding/json v1 round trip changed the value")
	}

	v2, err := json.Marshal(reading{Amount: amount})
	if err != nil {
		t.Fatal(err)
	}
	if string(v2) != want {
		t.Fatalf("encoding/json v2 = %s, want %s", v2, want)
	}
	var decodedV2 reading
	if err := json.Unmarshal(v2, &decodedV2); err != nil {
		t.Fatal(err)
	}
	if decodedV2.Amount.Cmp(amount) != 0 {
		t.Fatal("encoding/json v2 round trip changed the value")
	}

	var bad reading
	if err := jsonv1.Unmarshal([]byte(`{"amount":"1e"}`), &bad); !errors.Is(err, &Error{Code: CodeInvalidValue}) {
		t.Fatalf("malformed decimal error = %v, want invalid_value", err)
	}
	if err := jsonv1.Unmarshal([]byte(`{"amount":12.34}`), &bad); err == nil {
		t.Fatal("JSON number accepted where a string is required")
	}
}

func TestDecimalIsNotComparableWithOperator(t *testing.T) {
	if reflect.TypeOf(Decimal{}).Comparable() {
		t.Fatal("Decimal must not be comparable with ==, it would compare pointers")
	}
}

func TestDecimalEqualAndCmp(t *testing.T) {
	cases := []struct {
		left, right string
		cmp         int
	}{
		{"1", "1", 0},
		{"1.0", "1", 0},
		{"0", "-0", 0},
		{"1e3", "1000", 0},
		{"-1", "1", -1},
		{"1.5", "1.25", 1},
		{"1e-1500", "0", 1},
		{"-1e1500", "1e-1500", -1},
	}
	for _, c := range cases {
		left, err := ParseDecimal(c.left)
		if err != nil {
			t.Fatal(err)
		}
		right, err := ParseDecimal(c.right)
		if err != nil {
			t.Fatal(err)
		}
		if got := left.Cmp(right); got != c.cmp {
			t.Errorf("Cmp(%s, %s) = %d, want %d", c.left, c.right, got, c.cmp)
		}
		if got := right.Cmp(left); got != -c.cmp {
			t.Errorf("Cmp(%s, %s) = %d, want %d", c.right, c.left, got, -c.cmp)
		}
		if got := left.Equal(right); got != (c.cmp == 0) {
			t.Errorf("Equal(%s, %s) = %v, want %v", c.left, c.right, got, c.cmp == 0)
		}
	}
	var zero Decimal
	if !zero.Equal(Decimal{}) || zero.Cmp(Decimal{}) != 0 {
		t.Fatal("zero values are not equal")
	}
}

package quantity

import (
	"errors"
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
		{"1e-100000", UnitLengthKilometre},
		{"1e100000", UnitLengthMillimetre},
		{"-7e-99999", UnitLengthKilometre},
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

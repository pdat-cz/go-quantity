package quantity

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func TestFloatPathsRoundOnce(t *testing.T) {
	boiling, err := New("100", UnitTemperatureCelsius)
	if err != nil {
		t.Fatal(err)
	}
	fahrenheit, err := boiling.InFloat64(UnitTemperatureFahrenheit)
	if err != nil {
		t.Fatal(err)
	}
	if fahrenheit != 212 {
		t.Fatalf("100 °C in °F = %.17g, want exactly 212", fahrenheit)
	}

	tenth, err := NewFloat64(0.1, UnitElectricCurrentMilliampere)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := tenth.Decimal().String(), "0.0001"; got != want {
		t.Fatalf("0.1 mA stored as %q A, want %q", got, want)
	}

	body, err := NewFloat64(98.6, UnitTemperatureFahrenheit)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := body.Decimal().String(), "310.15"; got != want {
		t.Fatalf("98.6 °F stored as %q K, want %q", got, want)
	}
}

func TestFloatPathsDoNotOverflowOnIntermediate(t *testing.T) {
	huge, err := New("1e309", UnitLengthMetre)
	if err != nil {
		t.Fatal(err)
	}
	kilometres, err := huge.InFloat64(UnitLengthKilometre)
	if err != nil {
		t.Fatal(err)
	}
	if kilometres != 1e306 {
		t.Fatalf("1e309 m in km = %g, want 1e306", kilometres)
	}
	if _, err := huge.InFloat64(UnitLengthMetre); !errors.Is(err, &Error{Code: CodeInvalidValue}) {
		t.Fatalf("1e309 m as float64 error = %v, want invalid_value", err)
	}

	tiny, err := New("1e-324", UnitLengthMetre)
	if err != nil {
		t.Fatal(err)
	}
	millimetres, err := tiny.InFloat64(UnitLengthMillimetre)
	if err != nil {
		t.Fatal(err)
	}
	if millimetres == 0 || math.Abs(millimetres-1e-321)/1e-321 > 1e-2 {
		t.Fatalf("1e-324 m in mm = %g, want about 1e-321", millimetres)
	}
}

func TestNewFloat64RejectsNonFinite(t *testing.T) {
	for _, amount := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := NewFloat64(amount, UnitLengthMetre); !errors.Is(err, &Error{Code: CodeInvalidValue}) {
			t.Errorf("NewFloat64(%v) error = %v, want invalid_value", amount, err)
		}
	}
}

func TestNewTransformBoundsAndValidatesRationals(t *testing.T) {
	huge := strings.Repeat("9", 1_000_000)
	invalid := [][4]string{
		{huge, "1", "0", "1"},
		{"1", huge, "0", "1"},
		{"1", "1", huge, "1"},
		{"1", "1", "0", huge},
		{"0", "1", "0", "1"},
		{"1", "0", "0", "1"},
		{"+1", "1", "0", "1"},
		{"1", "1", "0x1", "1"},
		{"", "1", "0", "1"},
	}
	for _, c := range invalid {
		start := time.Now()
		if _, err := NewTransform(c[0], c[1], c[2], c[3]); !errors.Is(err, &Error{Code: CodeInvalidDefinition}) {
			t.Errorf("NewTransform(%q,%q,%q,%q) error = %v, want invalid_definition", c[0], c[1], c[2], c[3], err)
		}
		if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
			t.Errorf("rejection of %d-byte input took %s", len(c[0])+len(c[1])+len(c[2])+len(c[3]), elapsed)
		}
	}
	transform, err := NewTransform("-5", "-9", "0", "-1")
	if err != nil {
		t.Fatal(err)
	}
	if n, d := transform.Scale(); n != "5" || d != "9" {
		t.Fatalf("Scale() = %s/%s, want 5/9", n, d)
	}
}

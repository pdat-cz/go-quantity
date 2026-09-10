package quantity

import (
	"errors"
	"math"
	"reflect"
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

func TestTransformIsNotComparableWithOperator(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(Transform{}), reflect.TypeOf(UnitDefinition{})} {
		if typ.Comparable() {
			t.Errorf("%s must not be comparable with ==, it would compare pointers", typ)
		}
	}
}

func TestTransformEqual(t *testing.T) {
	half := mustTransform("1", "2", "0", "1")
	if !half.Equal(mustTransform("2", "4", "0", "1")) {
		t.Fatal("1/2 and 2/4 should be equal")
	}
	if half.Equal(mustTransform("1", "2", "1", "1")) {
		t.Fatal("different offsets compared equal")
	}
	if !IdentityTransform().Equal(mustTransform("1", "1", "0", "1")) {
		t.Fatal("identity should equal 1/1 + 0")
	}
	var zero Transform
	if zero.Equal(IdentityTransform()) || !zero.Equal(Transform{}) {
		t.Fatal("zero transform equality is wrong")
	}
	if got := StandardCatalog.Units(KindLength); !got[0].ToReference().Equal(mustTransform("1", "100", "0", "1")) {
		t.Fatalf("first length unit %s has unexpected transform", got[0].ID())
	}
}

// TestEveryStandardUnitInvertsExactly checks reference = value*scale + offset
// and back for every built-in unit. Units whose transform has a non-decimal
// offset (Fahrenheit) may legitimately report CodeInexact on the way in; if
// the value went in, it must come back out unchanged.
func TestEveryStandardUnitInvertsExactly(t *testing.T) {
	samples := []string{"0", "1", "-1", "0.5", "1625", "-273.15", "1e-10", "123456789.987654321"}
	kinds := []Kind{KindElectricCurrent, KindLength, KindTemperature, KindTemperatureDifference}
	checked := 0
	for _, kind := range kinds {
		for _, unit := range StandardCatalog.Units(kind) {
			for _, text := range samples {
				amount, err := ParseDecimal(text)
				if err != nil {
					t.Fatal(err)
				}
				value, err := StandardCatalog.New(amount, unit.ID())
				if errors.Is(err, &Error{Code: CodeInexact}) {
					continue
				}
				if err != nil {
					t.Fatalf("New(%s, %s): %v", text, unit.ID(), err)
				}
				back, err := value.In(unit.ID())
				if err != nil {
					t.Fatalf("In(%s) of %s: %v", unit.ID(), text, err)
				}
				if !back.Equal(amount) {
					t.Errorf("%s %s round-tripped as %s", text, unit.ID(), back)
				}
				checked++
			}
		}
	}
	if checked < 80 {
		t.Fatalf("only %d round trips checked, the sample set is too small", checked)
	}
}

func TestKnownConversions(t *testing.T) {
	cases := []struct {
		amount   string
		from, to UnitID
		want     string
	}{
		{"1", UnitLengthKilometre, UnitLengthMillimetre, "1000000"},
		{"2.54", UnitLengthCentimetre, UnitLengthMetre, "0.0254"},
		{"-40", UnitTemperatureCelsius, UnitTemperatureFahrenheit, "-40"},
		{"0", UnitTemperatureKelvin, UnitTemperatureCelsius, "-273.15"},
		{"373.15", UnitTemperatureKelvin, UnitTemperatureFahrenheit, "212"},
		{"9", UnitTemperatureDifferenceFahrenheit, UnitTemperatureDifferenceKelvin, "5"},
		{"1", UnitElectricCurrentKiloampere, UnitElectricCurrentMicroampere, "1000000000"},
	}
	for _, c := range cases {
		value, err := New(c.amount, c.from)
		if err != nil {
			t.Fatalf("New(%s, %s): %v", c.amount, c.from, err)
		}
		got, err := value.In(c.to)
		if err != nil {
			t.Fatalf("%s %s in %s: %v", c.amount, c.from, c.to, err)
		}
		if got.String() != c.want {
			t.Errorf("%s %s in %s = %s, want %s", c.amount, c.from, c.to, got, c.want)
		}
	}
}

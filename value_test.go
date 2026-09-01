package quantity

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
)

func TestValueUsesReferenceUnit(t *testing.T) {
	value, err := New("1625", UnitElectricCurrentMilliampere)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := value.Decimal().String(), "1.625"; got != want {
		t.Fatalf("Decimal() = %q, want %q", got, want)
	}
	if got, want := value.Unit(), UnitElectricCurrentAmpere; got != want {
		t.Fatalf("Unit() = %q, want %q", got, want)
	}
	if got, want := value.Kind(), KindElectricCurrent; got != want {
		t.Fatalf("Kind() = %q, want %q", got, want)
	}
	milliamperes, err := value.In(UnitElectricCurrentMilliampere)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := milliamperes.String(), "1625"; got != want {
		t.Fatalf("value in mA = %q, want %q", got, want)
	}
}

func TestTemperatureAndDifferenceAreDistinctKinds(t *testing.T) {
	temperature, err := New("0", UnitTemperatureCelsius)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := temperature.Decimal().String(), "273.15"; got != want {
		t.Fatalf("0 °C = %q K, want %q", got, want)
	}
	if _, err := temperature.In(UnitTemperatureDifferenceCelsius); !errors.Is(err, &Error{Code: CodeKindMismatch}) {
		t.Fatalf("cross-kind conversion error = %v, want kind mismatch", err)
	}

	difference, err := New("1", UnitTemperatureDifferenceCelsius)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := difference.Decimal().String(), "1"; got != want {
		t.Fatalf("1 °C difference = %q K, want %q", got, want)
	}
}

func TestExactAndFloatBoundaries(t *testing.T) {
	if _, err := New("1", UnitTemperatureFahrenheit); !errors.Is(err, &Error{Code: CodeInexact}) {
		t.Fatalf("exact Fahrenheit conversion error = %v, want inexact", err)
	}
	value, err := NewFloat64(1, UnitTemperatureFahrenheit)
	if err != nil {
		t.Fatal(err)
	}
	converted, err := value.InFloat64(UnitTemperatureFahrenheit)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(converted-1) > 1e-12 {
		t.Fatalf("float round trip = %.17g, want 1", converted)
	}
}

func TestValueJSONIsSmallAndNormalJSON(t *testing.T) {
	input := []byte(" { \n  \"unit\" : \"electric_current.milliampere\", \n  \"value\" : \"1625\" \n } ")
	value, err := StandardCatalog.ParseJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"value":"1.625","unit":"electric_current.ampere"}`
	if got := string(encoded); got != want {
		t.Fatalf("JSON = %s, want %s", got, want)
	}
}

func TestValueJSONRejectsUnknownFields(t *testing.T) {
	_, err := StandardCatalog.ParseJSON([]byte(`{"value":"1","unit":"electric_current.ampere","extra":true}`))
	if err == nil {
		t.Fatal("unknown field unexpectedly accepted")
	}
}

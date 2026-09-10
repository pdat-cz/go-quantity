package quantity

import (
	jsonv1 "encoding/json"
	json "encoding/json/v2"
	"errors"
	"math"
	"strings"
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

func TestDataInSerializesSelectedUnit(t *testing.T) {
	value, err := New("1.625", UnitElectricCurrentAmpere)
	if err != nil {
		t.Fatal(err)
	}
	data, err := value.DataIn(UnitElectricCurrentMilliampere)
	if err != nil {
		t.Fatal(err)
	}
	if data.Value != "1625" || data.Unit != UnitElectricCurrentMilliampere {
		t.Fatalf("DataIn(mA) = %#v", data)
	}
	decoded, err := StandardCatalog.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Decimal().Cmp(value.Decimal()) != 0 {
		t.Fatal("Data round trip changed the value")
	}
}

func TestJSONV2RejectsAmbiguousInput(t *testing.T) {
	invalid := [][]byte{
		[]byte(`{"value":"1","unit":"electric_current.ampere","extra":true}`),
		[]byte(`{"value":"1","value":"2","unit":"electric_current.ampere"}`),
		[]byte(`{"Value":"1","unit":"electric_current.ampere"}`),
	}
	for _, input := range invalid {
		var data Data
		if err := json.Unmarshal(input, &data); err == nil {
			t.Errorf("ambiguous JSON unexpectedly accepted: %s", input)
		}
	}
}

func TestEncodingJSONV1Compatibility(t *testing.T) {
	value, err := New("1.625", UnitElectricCurrentAmpere)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := jsonv1.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Value
	if err := jsonv1.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Decimal().Cmp(value.Decimal()) != 0 || decoded.Unit() != value.Unit() {
		t.Fatal("encoding/json v1 round trip changed the value")
	}
}

func TestValueJSONRejectsUnknownFields(t *testing.T) {
	_, err := StandardCatalog.ParseJSON([]byte(`{"value":"1","unit":"electric_current.ampere","extra":true}`))
	if err == nil {
		t.Fatal("unknown field unexpectedly accepted")
	}
}

func TestValueJSONRoundTripsAtLimits(t *testing.T) {
	longest := UnitID(strings.Repeat("k", 128) + "." + strings.Repeat("u", 63))
	kind, _ := longest.Kind()
	catalog, err := NewCatalog(
		[]KindDefinition{mustKind(kind, "Long", longest)},
		[]UnitDefinition{mustUnit(longest, kind, "long", "l", IdentityTransform())},
	)
	if err != nil {
		t.Fatal(err)
	}
	nines := strings.Repeat("9", maxDecimalDigits)
	for _, text := range []string{nines + "e1500", "-" + nines + "e-1500"} {
		amount, err := ParseDecimal(text)
		if err != nil {
			t.Fatal(err)
		}
		value, err := catalog.New(amount, longest)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := value.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}
		if len(encoded) > maxValueJSONBytes {
			t.Fatalf("encoded %d bytes, over the ParseJSON limit", len(encoded))
		}
		decoded, err := catalog.ParseJSON(encoded)
		if err != nil {
			t.Fatalf("ParseJSON(MarshalJSON()): %v", err)
		}
		if decoded.Decimal().Cmp(amount) != 0 {
			t.Fatal("JSON round trip changed the value")
		}
	}
}

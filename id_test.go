package quantity

import (
	"errors"
	"strings"
	"testing"
)

func TestUnitIDIsScopedByKind(t *testing.T) {
	temperature, err := ParseUnitID("temperature.celsius")
	if err != nil {
		t.Fatal(err)
	}
	difference, err := ParseUnitID("temperature_difference.celsius")
	if err != nil {
		t.Fatal(err)
	}
	if temperature == difference {
		t.Fatal("scoped UnitIDs must be distinct")
	}
	if got, _ := temperature.Kind(); got != KindTemperature {
		t.Fatalf("Kind() = %q, want %q", got, KindTemperature)
	}
}

func TestUnitIDRejectsDerivedOrAmbiguousForm(t *testing.T) {
	invalid := []string{"ampere", "electric_current_ampere", "ElectricCurrent.ampere", "electric_current.ampere.extra"}
	for _, input := range invalid {
		if _, err := ParseUnitID(input); err == nil {
			t.Errorf("ParseUnitID(%q) unexpectedly succeeded", input)
		}
	}
}

func FuzzParseUnitID(f *testing.F) {
	for _, seed := range []string{"length.metre", "a.b", "a", "a.b.c", "A.b", "a.", ".b", "", "a\n.b"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		id, err := ParseUnitID(text)
		if err != nil {
			if !errors.Is(err, &Error{Code: CodeInvalidID}) {
				t.Fatalf("unexpected error %v", err)
			}
			return
		}
		kind, err := id.Kind()
		if err != nil || !kind.Valid() || !id.Valid() {
			t.Fatalf("accepted id %q is not self-consistent", text)
		}
		if strings.Count(text, ".") != 1 || strings.ContainsAny(text, "\n\r\x00 ") {
			t.Fatalf("accepted malformed id %q", text)
		}
	})
}

package quantity

import "testing"

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

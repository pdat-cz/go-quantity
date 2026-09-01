package quantity

import "testing"

func TestReferenceUnitMustUseIdentityTransform(t *testing.T) {
	kind := mustKind("example", "Example", "example.reference")
	unit := mustUnit("example.reference", "example", "reference", "r", mustTransform("2", "1", "0", "1"))
	if _, err := NewCatalog([]KindDefinition{kind}, []UnitDefinition{unit}); err == nil {
		t.Fatal("non-identity ReferenceUnit unexpectedly accepted")
	}
}

func TestUnitsReturnsOnlyRequestedKind(t *testing.T) {
	units := StandardCatalog.Units(KindElectricCurrent)
	if len(units) != 4 {
		t.Fatalf("len(Units(electric_current)) = %d, want 4", len(units))
	}
	for _, unit := range units {
		if unit.Kind() != KindElectricCurrent {
			t.Fatalf("unit %q has kind %q", unit.ID(), unit.Kind())
		}
	}
}

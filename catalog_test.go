package quantity

import (
	"errors"
	"testing"
)

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

func TestNewCatalogRejectsEveryInvalidShape(t *testing.T) {
	kind := mustKind("example", "Example", "example.reference")
	reference := mustUnit("example.reference", "example", "reference", "r", IdentityTransform())
	double := mustUnit("example.double", "example", "double", "d", mustTransform("2", "1", "0", "1"))
	cases := map[string]struct {
		kinds []KindDefinition
		units []UnitDefinition
		code  Code
	}{
		"zero kind definition":   {[]KindDefinition{{}}, nil, CodeInvalidDefinition},
		"zero unit definition":   {[]KindDefinition{kind}, []UnitDefinition{reference, {}}, CodeInvalidDefinition},
		"duplicate kind":         {[]KindDefinition{kind, kind}, []UnitDefinition{reference}, CodeDuplicateKind},
		"duplicate unit":         {[]KindDefinition{kind}, []UnitDefinition{reference, reference}, CodeDuplicateUnit},
		"unit of unknown kind":   {nil, []UnitDefinition{reference}, CodeUnknownKind},
		"missing reference unit": {[]KindDefinition{kind}, []UnitDefinition{double}, CodeUnknownUnit},
		"non-identity reference": {[]KindDefinition{mustKind("example", "Example", "example.double")}, []UnitDefinition{double}, CodeInvalidDefinition},
	}
	for name, c := range cases {
		_, err := NewCatalog(c.kinds, c.units)
		if !errors.Is(err, &Error{Code: c.code}) {
			t.Errorf("%s: error = %v, want %s", name, err, c.code)
		}
	}
	if _, err := NewCatalog([]KindDefinition{kind}, []UnitDefinition{reference, double}); err != nil {
		t.Fatalf("valid catalog rejected: %v", err)
	}
}

func TestDefinitionConstructorsValidate(t *testing.T) {
	if _, err := NewKindDefinition("Example", "Example", "example.reference"); !errors.Is(err, &Error{Code: CodeInvalidDefinition}) {
		t.Errorf("uppercase kind error = %v", err)
	}
	if _, err := NewKindDefinition("example", " padded ", "example.reference"); !errors.Is(err, &Error{Code: CodeInvalidDefinition}) {
		t.Errorf("padded name error = %v", err)
	}
	if _, err := NewKindDefinition("example", "Example", "other.reference"); !errors.Is(err, &Error{Code: CodeKindMismatch}) {
		t.Errorf("reference of other kind error = %v", err)
	}
	if _, err := NewUnitDefinition("example.unit", "other", "unit", "u", IdentityTransform()); !errors.Is(err, &Error{Code: CodeKindMismatch}) {
		t.Errorf("unit id of other kind error = %v", err)
	}
	if _, err := NewUnitDefinition("example.unit", "example", "unit", "u\x00", IdentityTransform()); !errors.Is(err, &Error{Code: CodeInvalidDefinition}) {
		t.Errorf("control char in symbol error = %v", err)
	}
	if _, err := NewUnitDefinition("example.unit", "example", "unit", "u", Transform{}); !errors.Is(err, &Error{Code: CodeInvalidDefinition}) {
		t.Errorf("zero transform error = %v", err)
	}
	var nilCatalog *Catalog
	if _, ok := nilCatalog.Unit(UnitLengthMetre); ok {
		t.Error("nil catalog returned a unit")
	}
	if _, err := nilCatalog.New(Decimal{}, UnitLengthMetre); !errors.Is(err, &Error{Code: CodeInvalidDefinition}) {
		t.Errorf("nil catalog New error = %v", err)
	}
}

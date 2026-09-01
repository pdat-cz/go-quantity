package quantity

import "testing"

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

package quantity

// The first built-in vertical slice. The definitions will be generated from
// the public OQS catalog once that catalog is published.

const (
	KindElectricCurrent       Kind = "electric_current"
	KindLength                Kind = "length"
	KindTemperature           Kind = "temperature"
	KindTemperatureDifference Kind = "temperature_difference"

	UnitElectricCurrentAmpere      UnitID = "electric_current.ampere"
	UnitElectricCurrentMilliampere UnitID = "electric_current.milliampere"
	UnitElectricCurrentMicroampere UnitID = "electric_current.microampere"
	UnitElectricCurrentKiloampere  UnitID = "electric_current.kiloampere"

	UnitLengthMetre      UnitID = "length.metre"
	UnitLengthMillimetre UnitID = "length.millimetre"
	UnitLengthCentimetre UnitID = "length.centimetre"
	UnitLengthKilometre  UnitID = "length.kilometre"

	UnitTemperatureKelvin     UnitID = "temperature.kelvin"
	UnitTemperatureCelsius    UnitID = "temperature.celsius"
	UnitTemperatureFahrenheit UnitID = "temperature.fahrenheit"

	UnitTemperatureDifferenceKelvin     UnitID = "temperature_difference.kelvin"
	UnitTemperatureDifferenceCelsius    UnitID = "temperature_difference.celsius"
	UnitTemperatureDifferenceFahrenheit UnitID = "temperature_difference.fahrenheit"
)

// StandardCatalog is the built-in OQS catalog. Its definitions are immutable.
var StandardCatalog = mustStandardCatalog()

func mustStandardCatalog() *Catalog {
	kinds := []KindDefinition{
		mustKind(KindElectricCurrent, "Electric current", UnitElectricCurrentAmpere),
		mustKind(KindLength, "Length", UnitLengthMetre),
		mustKind(KindTemperature, "Temperature", UnitTemperatureKelvin),
		mustKind(KindTemperatureDifference, "Temperature difference", UnitTemperatureDifferenceKelvin),
	}
	identity := IdentityTransform()
	units := []UnitDefinition{
		mustUnit(UnitElectricCurrentAmpere, KindElectricCurrent, "ampere", "A", identity),
		mustUnit(UnitElectricCurrentMilliampere, KindElectricCurrent, "milliampere", "mA", mustTransform("1", "1000", "0", "1")),
		mustUnit(UnitElectricCurrentMicroampere, KindElectricCurrent, "microampere", "µA", mustTransform("1", "1000000", "0", "1")),
		mustUnit(UnitElectricCurrentKiloampere, KindElectricCurrent, "kiloampere", "kA", mustTransform("1000", "1", "0", "1")),

		mustUnit(UnitLengthMetre, KindLength, "metre", "m", identity),
		mustUnit(UnitLengthMillimetre, KindLength, "millimetre", "mm", mustTransform("1", "1000", "0", "1")),
		mustUnit(UnitLengthCentimetre, KindLength, "centimetre", "cm", mustTransform("1", "100", "0", "1")),
		mustUnit(UnitLengthKilometre, KindLength, "kilometre", "km", mustTransform("1000", "1", "0", "1")),

		mustUnit(UnitTemperatureKelvin, KindTemperature, "kelvin", "K", identity),
		mustUnit(UnitTemperatureCelsius, KindTemperature, "degree Celsius", "°C", mustTransform("1", "1", "27315", "100")),
		mustUnit(UnitTemperatureFahrenheit, KindTemperature, "degree Fahrenheit", "°F", mustTransform("5", "9", "45967", "180")),

		mustUnit(UnitTemperatureDifferenceKelvin, KindTemperatureDifference, "kelvin", "K", identity),
		mustUnit(UnitTemperatureDifferenceCelsius, KindTemperatureDifference, "degree Celsius", "°C", identity),
		mustUnit(UnitTemperatureDifferenceFahrenheit, KindTemperatureDifference, "degree Fahrenheit", "°F", mustTransform("5", "9", "0", "1")),
	}
	catalog, err := NewCatalog(kinds, units)
	if err != nil {
		panic(err)
	}
	return catalog
}

func mustKind(id Kind, name string, referenceUnit UnitID) KindDefinition {
	definition, err := NewKindDefinition(id, name, referenceUnit)
	if err != nil {
		panic(err)
	}
	return definition
}

func mustUnit(id UnitID, kind Kind, name, symbol string, transform Transform) UnitDefinition {
	definition, err := NewUnitDefinition(id, kind, name, symbol, transform)
	if err != nil {
		panic(err)
	}
	return definition
}

func mustTransform(scaleNumerator, scaleDenominator, offsetNumerator, offsetDenominator string) Transform {
	transform, err := NewTransform(scaleNumerator, scaleDenominator, offsetNumerator, offsetDenominator)
	if err != nil {
		panic(err)
	}
	return transform
}

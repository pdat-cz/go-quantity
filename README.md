# go-quantity

Go implementation of [Open Quantity Schema](https://oqs.microon.org/).

OQS owns the public model and unit catalogs. `go-quantity` implements that
model for Go; it is not a second source of unit definitions.

## Vocabulary

| OQS | Go |
| --- | --- |
| `QuantityValue` | `quantity.Value` |
| `QuantityKind` | `quantity.Kind` |
| `Unit` | `quantity.UnitDefinition` |
| `UnitID` | `quantity.UnitID` |
| `ReferenceUnit` | `KindDefinition.ReferenceUnit()` |

`Measurement` is measurement process or context and is outside the minimal
model. `Dimension` is not part of the computational core.

## Example

```go
current, err := quantity.New("1625", quantity.UnitElectricCurrentMilliampere)
if err != nil {
	return err
}

fmt.Println(current.Decimal()) // 1.625
fmt.Println(current.Unit())    // electric_current.ampere

milliamperes, err := current.In(quantity.UnitElectricCurrentMilliampere)
```

Values are stored as an exact `Decimal` in their `ReferenceUnit`. Unit
conversions use exact rational affine transforms. A conversion that cannot be
represented by a finite decimal returns `CodeInexact`; `NewFloat64` and
`InFloat64` provide an explicit approximation boundary.

The minimal JSON representation is:

```json
{
  "value": "1.625",
  "unit": "electric_current.ampere"
}
```

## Current scope

The first vertical slice contains electric current, length, temperature, and
temperature difference. These built-in definitions are temporary and will be
generated from the public OQS catalog.

Catalog signatures, extension namespaces, dimension algebra, compound-unit
inference, localization, and generated convenience APIs are intentionally
deferred until the minimal model is stable.

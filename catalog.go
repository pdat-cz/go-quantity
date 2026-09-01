package quantity

import (
	"errors"
	"sort"
	"strings"
	"unicode/utf8"
)

// KindDefinition describes an OQS QuantityKind and its ReferenceUnit.
type KindDefinition struct {
	id            Kind
	name          string
	referenceUnit UnitID
}

// NewKindDefinition constructs a validated QuantityKind definition.
func NewKindDefinition(id Kind, name string, referenceUnit UnitID) (KindDefinition, error) {
	if !id.Valid() || !referenceUnit.Valid() || !validLabel(name) {
		return KindDefinition{}, &Error{Code: CodeInvalidDefinition, Op: "new kind definition", Kind: id}
	}
	referenceKind, _ := referenceUnit.Kind()
	if referenceKind != id {
		return KindDefinition{}, &Error{Code: CodeKindMismatch, Op: "new kind definition", Kind: id, Unit: referenceUnit}
	}
	return KindDefinition{id: id, name: name, referenceUnit: referenceUnit}, nil
}

func (d KindDefinition) ID() Kind              { return d.id }
func (d KindDefinition) Name() string          { return d.name }
func (d KindDefinition) ReferenceUnit() UnitID { return d.referenceUnit }

// UnitDefinition describes an OQS Unit and its exact conversion to the
// QuantityKind's ReferenceUnit.
type UnitDefinition struct {
	id          UnitID
	kind        Kind
	name        string
	symbol      string
	toReference Transform
}

// NewUnitDefinition constructs a validated Unit definition.
func NewUnitDefinition(id UnitID, kind Kind, name, symbol string, toReference Transform) (UnitDefinition, error) {
	if !id.Valid() || !kind.Valid() || !validLabel(name) || !validLabel(symbol) || !toReference.valid() {
		return UnitDefinition{}, &Error{Code: CodeInvalidDefinition, Op: "new unit definition", Kind: kind, Unit: id}
	}
	idKind, _ := id.Kind()
	if idKind != kind {
		return UnitDefinition{}, &Error{Code: CodeKindMismatch, Op: "new unit definition", Kind: kind, Unit: id}
	}
	return UnitDefinition{id: id, kind: kind, name: name, symbol: symbol, toReference: toReference}, nil
}

func (d UnitDefinition) ID() UnitID             { return d.id }
func (d UnitDefinition) Kind() Kind             { return d.kind }
func (d UnitDefinition) Name() string           { return d.name }
func (d UnitDefinition) Symbol() string         { return d.symbol }
func (d UnitDefinition) ToReference() Transform { return d.toReference }

// Catalog is an immutable set of QuantityKind and Unit definitions.
type Catalog struct {
	kinds map[Kind]KindDefinition
	units map[UnitID]UnitDefinition
}

// NewCatalog validates definitions and creates an immutable catalog.
func NewCatalog(kinds []KindDefinition, units []UnitDefinition) (*Catalog, error) {
	catalog := &Catalog{
		kinds: make(map[Kind]KindDefinition, len(kinds)),
		units: make(map[UnitID]UnitDefinition, len(units)),
	}
	for _, definition := range kinds {
		validated, err := NewKindDefinition(definition.id, definition.name, definition.referenceUnit)
		if err != nil {
			return nil, err
		}
		if _, exists := catalog.kinds[validated.id]; exists {
			return nil, &Error{Code: CodeDuplicateKind, Op: "new catalog", Kind: validated.id}
		}
		catalog.kinds[validated.id] = validated
	}
	for _, definition := range units {
		validated, err := NewUnitDefinition(definition.id, definition.kind, definition.name, definition.symbol, definition.toReference)
		if err != nil {
			return nil, err
		}
		if _, exists := catalog.kinds[validated.kind]; !exists {
			return nil, &Error{Code: CodeUnknownKind, Op: "new catalog", Kind: validated.kind, Unit: validated.id}
		}
		if _, exists := catalog.units[validated.id]; exists {
			return nil, &Error{Code: CodeDuplicateUnit, Op: "new catalog", Unit: validated.id}
		}
		catalog.units[validated.id] = validated
	}
	for kind, definition := range catalog.kinds {
		reference, exists := catalog.units[definition.referenceUnit]
		if !exists {
			return nil, &Error{Code: CodeUnknownUnit, Op: "new catalog", Kind: kind, Unit: definition.referenceUnit}
		}
		if reference.kind != kind {
			return nil, &Error{Code: CodeKindMismatch, Op: "new catalog", Kind: kind, Unit: reference.id}
		}
		if !reference.toReference.IsIdentity() {
			return nil, &Error{Code: CodeInvalidDefinition, Op: "new catalog", Kind: kind, Unit: reference.id, Err: errors.New("ReferenceUnit must use the identity transform")}
		}
	}
	return catalog, nil
}

// Kind returns a QuantityKind definition.
func (c *Catalog) Kind(id Kind) (KindDefinition, bool) {
	if c == nil {
		return KindDefinition{}, false
	}
	definition, ok := c.kinds[id]
	return definition, ok
}

// Unit returns a Unit definition.
func (c *Catalog) Unit(id UnitID) (UnitDefinition, bool) {
	if c == nil {
		return UnitDefinition{}, false
	}
	definition, ok := c.units[id]
	return definition, ok
}

// Units returns all units belonging to kind, ordered by UnitID.
func (c *Catalog) Units(kind Kind) []UnitDefinition {
	if c == nil {
		return nil
	}
	definitions := make([]UnitDefinition, 0)
	for _, definition := range c.units {
		if definition.kind == kind {
			definitions = append(definitions, definition)
		}
	}
	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].id < definitions[j].id
	})
	return definitions
}

func validLabel(text string) bool {
	if text == "" || len(text) > 256 || !utf8.ValidString(text) || strings.TrimSpace(text) != text {
		return false
	}
	for _, char := range text {
		if char < 0x20 || char == 0x7f {
			return false
		}
	}
	return true
}

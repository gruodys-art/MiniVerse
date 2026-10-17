// Package chem is a small simulation of chemical parcels.
//
// A parcel is one material sitting somewhere on a plane, with an amount in
// moles and a temperature in kelvin. Each tick, parcels that are close enough
// share heat and may react. Reactions come only from the catalog in this file.
// The simulation does not invent new chemistry.
package chem

import "fmt"

// SpeciesID is the short name used to look a material up in the catalog.
type SpeciesID string

const (
	Water         SpeciesID = "water"
	Ice           SpeciesID = "ice"
	Steam         SpeciesID = "steam"
	Salt          SpeciesID = "salt"
	DissolvedSalt SpeciesID = "dissolved_salt"
	Acid          SpeciesID = "acid"
	Base          SpeciesID = "base"
	Iron          SpeciesID = "iron"
	Oxygen        SpeciesID = "oxygen"
	IronOxide     SpeciesID = "iron_oxide"
)

// Species is one material. Atoms is the count of each element in a single
// formula unit, which is what makes a mole of that material.
type Species struct {
	ID           SpeciesID
	Name         string
	Atoms        map[string]int
	MolarMass    float64 // grams per mole
	HeatCapacity float64 // joules per mole per kelvin
}

// Term is one side of a reaction: Count moles of Species take part per extent.
// An extent of 1 means the reaction happens once, as written.
type Term struct {
	Species SpeciesID
	Count   float64
}

// Reaction is one allowed interaction. Needs is a material that must be
// nearby but is not consumed, such as water when salt dissolves.
// Enthalpy is joules per extent. A negative value releases heat.
type Reaction struct {
	ID          string
	Reactants   []Term
	Products    []Term
	Needs       SpeciesID
	MaxDistance float64
	// Ea is the activation energy in joules per mole.
	// PreFactor is the Arrhenius multiplier, in extent per second.
	Ea        float64
	PreFactor float64
	Enthalpy  float64
}

// Catalog is the closed list of materials and reactions.
type Catalog struct {
	Species   map[SpeciesID]Species
	Reactions []Reaction
}

// atomicMass is grams per mole of each element we use.
var atomicMass = map[string]float64{
	"H":  1.008,
	"O":  15.999,
	"Na": 22.990,
	"Cl": 35.45,
	"Fe": 55.845,
}

// DefaultCatalog returns the materials the sandbox knows about.
func DefaultCatalog() Catalog {
	species := []Species{
		{ID: Water, Name: "water", Atoms: map[string]int{"H": 2, "O": 1}, HeatCapacity: 75.3},
		{ID: Ice, Name: "ice", Atoms: map[string]int{"H": 2, "O": 1}, HeatCapacity: 37.7},
		{ID: Steam, Name: "steam", Atoms: map[string]int{"H": 2, "O": 1}, HeatCapacity: 36.5},
		{ID: Salt, Name: "salt", Atoms: map[string]int{"Na": 1, "Cl": 1}, HeatCapacity: 50.5},
		{ID: DissolvedSalt, Name: "dissolved salt", Atoms: map[string]int{"Na": 1, "Cl": 1}, HeatCapacity: 50.5},
		{ID: Acid, Name: "acid", Atoms: map[string]int{"H": 1, "Cl": 1}, HeatCapacity: 29.1},
		{ID: Base, Name: "base", Atoms: map[string]int{"Na": 1, "O": 1, "H": 1}, HeatCapacity: 59.5},
		{ID: Iron, Name: "iron", Atoms: map[string]int{"Fe": 1}, HeatCapacity: 25.1},
		{ID: Oxygen, Name: "oxygen", Atoms: map[string]int{"O": 2}, HeatCapacity: 29.4},
		{ID: IronOxide, Name: "iron oxide", Atoms: map[string]int{"Fe": 2, "O": 3}, HeatCapacity: 103.9},
	}

	byID := make(map[SpeciesID]Species, len(species))
	for _, s := range species {
		s.MolarMass = molarMass(s.Atoms)
		byID[s.ID] = s
	}

	return Catalog{
		Species: byID,
		Reactions: []Reaction{
			{
				ID:          "dissolve_salt",
				Reactants:   []Term{{Species: Salt, Count: 1}},
				Products:    []Term{{Species: DissolvedSalt, Count: 1}},
				Needs:       Water,
				MaxDistance: 2,
				Ea:          5000,
				PreFactor:   30,
				Enthalpy:    3800,
			},
			{
				ID:          "neutralize",
				Reactants:   []Term{{Species: Acid, Count: 1}, {Species: Base, Count: 1}},
				Products:    []Term{{Species: DissolvedSalt, Count: 1}, {Species: Water, Count: 1}},
				MaxDistance: 2,
				Ea:          20000,
				PreFactor:   100,
				Enthalpy:    -56000,
			},
			{
				ID:          "oxidize_iron",
				Reactants:   []Term{{Species: Iron, Count: 4}, {Species: Oxygen, Count: 3}},
				Products:    []Term{{Species: IronOxide, Count: 2}},
				MaxDistance: 2,
				Ea:          60000,
				PreFactor:   5,
				Enthalpy:    -1648400,
			},
		},
	}
}

// Must returns a species or panics. A missing id is a programming mistake.
func (c Catalog) Must(id SpeciesID) Species {
	s, ok := c.Species[id]
	if !ok {
		panic("unknown species " + string(id))
	}
	return s
}

// Validate checks that every reaction refers to known species and that the
// atoms on the left match the atoms on the right.
func (c Catalog) Validate() error {
	for _, rx := range c.Reactions {
		if err := c.checkShape(rx); err != nil {
			return err
		}
		if err := c.checkBalance(rx); err != nil {
			return err
		}
	}
	for id, s := range c.Species {
		if s.MolarMass <= 0 {
			return fmt.Errorf("species %s has no molar mass", id)
		}
		for el := range s.Atoms {
			if _, ok := atomicMass[el]; !ok {
				return fmt.Errorf("species %s uses unknown element %s", id, el)
			}
		}
	}
	return nil
}

func (c Catalog) checkShape(rx Reaction) error {
	switch {
	case len(rx.Reactants) == 1 && rx.Needs != "":
	case len(rx.Reactants) == 2 && rx.Needs == "":
	default:
		return fmt.Errorf("reaction %s must be a pair, or one reactant beside a needed neighbor", rx.ID)
	}
	for _, term := range append(append([]Term{}, rx.Reactants...), rx.Products...) {
		if term.Count <= 0 {
			return fmt.Errorf("reaction %s has a non-positive amount", rx.ID)
		}
		if _, ok := c.Species[term.Species]; !ok {
			return fmt.Errorf("reaction %s uses unknown species %s", rx.ID, term.Species)
		}
	}
	if rx.Needs != "" {
		if _, ok := c.Species[rx.Needs]; !ok {
			return fmt.Errorf("reaction %s needs unknown species %s", rx.ID, rx.Needs)
		}
	}
	return nil
}

func (c Catalog) checkBalance(rx Reaction) error {
	left := map[string]float64{}
	right := map[string]float64{}
	add := func(dst map[string]float64, terms []Term) {
		for _, term := range terms {
			for el, n := range c.Species[term.Species].Atoms {
				dst[el] += term.Count * float64(n)
			}
		}
	}
	add(left, rx.Reactants)
	add(right, rx.Products)
	for el, n := range left {
		if !near(n, right[el]) {
			return fmt.Errorf("reaction %s is unbalanced for %s: left %.4f right %.4f", rx.ID, el, n, right[el])
		}
	}
	for el, n := range right {
		if !near(n, left[el]) {
			return fmt.Errorf("reaction %s is unbalanced for %s: left %.4f right %.4f", rx.ID, el, left[el], n)
		}
	}
	return nil
}

func molarMass(formula map[string]int) float64 {
	var mass float64
	for el, n := range formula {
		mass += float64(n) * atomicMass[el]
	}
	return mass
}

func near(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-9
}

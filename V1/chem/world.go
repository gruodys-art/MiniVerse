package chem

import (
	"fmt"
	"math"
	"sort"
)

const (
	gasR            = 8.314 // joules per mole per kelvin
	heatConductance = 40.0  // joules per second per kelvin at full contact
	moleEpsilon     = 1e-9
	mergeDistance   = 0.25

	freezeK      = 273.15
	boilK        = 373.15
	latentFusion = 6008.0  // joules absorbed per mole when ice melts
	latentVapor  = 40660.0 // joules absorbed per mole when water boils
)

// Parcel is a pile of a single material.
type Parcel struct {
	ID      int
	Species SpeciesID
	Moles   float64
	X       float64
	Y       float64
	Temp    float64 // kelvin
}

// Reach is how far this pile can touch. A larger amount reaches farther.
func (p Parcel) Reach() float64 {
	if p.Moles <= 0 {
		return 0
	}
	return 0.5 + 0.5*math.Sqrt(p.Moles)
}

// ParcelView is a copy of a parcel plus its display name.
type ParcelView struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Species SpeciesID `json:"species"`
	Moles   float64   `json:"moles"`
	X       float64   `json:"x"`
	Y       float64   `json:"y"`
	Temp    float64   `json:"temp"`
}

// World is the collection of parcels the tick advances.
type World struct {
	// Ambient is the air temperature in kelvin.
	// AirLoss is how fast a pile sheds heat to that air, in joules per second
	// per kelvin for each unit of reach. Zero means the world is isolated.
	Ambient float64
	AirLoss float64

	parcels []*Parcel
	nextID  int
	cat     Catalog
}

// NewWorld returns an empty world loaded with the default catalog.
func NewWorld() *World {
	cat := DefaultCatalog()
	if err := cat.Validate(); err != nil {
		panic(err)
	}
	return &World{cat: cat, nextID: 1}
}

// Spawn places a new parcel. Coordinates use the same units as reaction distance.
func (w *World) Spawn(species SpeciesID, moles, x, y, temp float64) error {
	if _, ok := w.cat.Species[species]; !ok {
		return fmt.Errorf("unknown species %s", species)
	}
	if moles <= 0 {
		return fmt.Errorf("moles must be positive")
	}
	if temp <= 0 {
		return fmt.Errorf("temperature must be positive kelvin")
	}
	w.spawn(species, moles, x, y, temp)
	return nil
}

// Tick advances the world by dt seconds.
//
// Touching parcels share heat, then each pile sheds heat to the air.
// Reactions are proposed from those temperatures. If several reactions want
// the same pile, their claims are scaled so the pile cannot be spent twice.
// Water, ice, and steam then move across their freeze and boil points.
func (w *World) Tick(dt float64) {
	if dt <= 0 {
		return
	}
	w.shareHeat(dt)
	w.shedHeat(dt)
	proposals := w.propose(dt)
	w.limitExtents(proposals)
	for _, prop := range proposals {
		w.apply(prop)
	}
	w.phaseChange()
	w.coalesce()
	w.reap()
}

// Mass is the total grams of material currently in the world.
func (w *World) Mass() float64 {
	var mass float64
	for _, p := range w.parcels {
		mass += p.Moles * w.cat.Must(p.Species).MolarMass
	}
	return mass
}

// MolesOf sums every parcel of the given species.
func (w *World) MolesOf(id SpeciesID) float64 {
	var moles float64
	for _, p := range w.parcels {
		if p.Species == id {
			moles += p.Moles
		}
	}
	return moles
}

// Snapshot returns the parcels sorted by id, safe for the caller to keep.
func (w *World) Snapshot() []ParcelView {
	out := make([]ParcelView, len(w.parcels))
	for i, p := range w.parcels {
		out[i] = ParcelView{
			ID:      p.ID,
			Name:    w.cat.Must(p.Species).Name,
			Species: p.Species,
			Moles:   p.Moles,
			X:       p.X,
			Y:       p.Y,
			Temp:    p.Temp,
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (w *World) shareHeat(dt float64) {
	// Pair order follows parcel id, so the same world always heats the same way.
	for i := 0; i < len(w.parcels); i++ {
		for j := i + 1; j < len(w.parcels); j++ {
			w.sharePair(w.parcels[i], w.parcels[j], dt)
		}
	}
}

func (w *World) sharePair(a, b *Parcel, dt float64) {
	dist := distance(a, b)
	reach := a.Reach() + b.Reach()
	if reach <= 0 || dist > reach {
		return
	}
	ca := w.capacity(a)
	cb := w.capacity(b)
	if ca <= 0 || cb <= 0 {
		return
	}
	closeness := 1 - dist/reach
	q := heatConductance * (a.Temp - b.Temp) * closeness * dt
	// qEq is the energy that would make the two temperatures meet.
	qEq := (a.Temp - b.Temp) / (1/ca + 1/cb)
	q = clampTransfer(q, qEq)
	a.Temp -= q / ca
	b.Temp += q / cb
}

func (w *World) shedHeat(dt float64) {
	if w.AirLoss <= 0 {
		return
	}
	for _, p := range w.parcels {
		cap := w.capacity(p)
		if cap <= 0 {
			continue
		}
		q := w.AirLoss * p.Reach() * (p.Temp - w.Ambient) * dt
		// A tiny pile has little heat capacity. Never carry it past the air temperature.
		qMax := cap * (p.Temp - w.Ambient)
		q = clampTransfer(q, qMax)
		p.Temp -= q / cap
	}
}

func clampTransfer(q, qEq float64) float64 {
	if qEq >= 0 {
		if q < 0 {
			return 0
		}
		if q > qEq {
			return qEq
		}
		return q
	}
	if q > 0 {
		return 0
	}
	if q < qEq {
		return qEq
	}
	return q
}

type proposal struct {
	rx     Reaction
	aID    int
	bID    int
	extent float64
}

func (w *World) propose(dt float64) []proposal {
	var out []proposal
	for i := 0; i < len(w.parcels); i++ {
		for j := i + 1; j < len(w.parcels); j++ {
			a := w.parcels[i]
			b := w.parcels[j]
			dist := distance(a, b)
			for _, rx := range w.cat.Reactions {
				if !pairMatches(rx, a, b, dist) {
					continue
				}
				extent := w.extent(rx, a, b, dt)
				if extent <= 1e-12 {
					continue
				}
				out = append(out, proposal{rx: rx, aID: a.ID, bID: b.ID, extent: extent})
			}
		}
	}
	return out
}

func pairMatches(rx Reaction, a, b *Parcel, dist float64) bool {
	limit := math.Max(a.Reach()+b.Reach(), rx.MaxDistance)
	if dist > limit {
		return false
	}
	if len(rx.Reactants) == 1 {
		want := rx.Reactants[0].Species
		return (a.Species == want && b.Species == rx.Needs) || (b.Species == want && a.Species == rx.Needs)
	}
	left := rx.Reactants[0].Species
	right := rx.Reactants[1].Species
	return (a.Species == left && b.Species == right) || (a.Species == right && b.Species == left)
}

func (w *World) extent(rx Reaction, a, b *Parcel, dt float64) float64 {
	temp := 0.5 * (a.Temp + b.Temp)
	if temp < 1 {
		temp = 1
	}
	// k = A * exp(-Ea / RT). Hotter pairs react further in the same time.
	k := rx.PreFactor * math.Exp(-rx.Ea/(gasR*temp))
	limit := k * dt
	for _, term := range rx.Reactants {
		p := parcelWithSpecies(term.Species, a, b)
		if p == nil || term.Count <= 0 {
			return 0
		}
		available := p.Moles / term.Count
		if available < limit {
			limit = available
		}
	}
	if limit < 0 {
		return 0
	}
	return limit
}

// limitExtents shrinks proposals that share a parcel so consumption cannot
// exceed the moles that parcel actually has.
func (w *World) limitExtents(props []proposal) {
	byID := w.index()
	claim := map[int]float64{}
	for _, prop := range props {
		a := byID[prop.aID]
		b := byID[prop.bID]
		for _, term := range prop.rx.Reactants {
			p := parcelWithSpecies(term.Species, a, b)
			if p == nil {
				continue
			}
			claim[p.ID] += prop.extent * term.Count
		}
	}
	scale := map[int]float64{}
	for id, wanted := range claim {
		have := byID[id].Moles
		if wanted > have && wanted > 0 {
			scale[id] = have / wanted
		} else {
			scale[id] = 1
		}
	}
	for i := range props {
		factor := 1.0
		a := byID[props[i].aID]
		b := byID[props[i].bID]
		for _, term := range props[i].rx.Reactants {
			p := parcelWithSpecies(term.Species, a, b)
			if p != nil && scale[p.ID] < factor {
				factor = scale[p.ID]
			}
		}
		props[i].extent *= factor
	}
}

func (w *World) apply(prop proposal) {
	if prop.extent <= 0 {
		return
	}
	byID := w.index()
	a := byID[prop.aID]
	b := byID[prop.bID]
	if a == nil || b == nil {
		return
	}

	// Capture thermal energy first. A negative enthalpy releases heat, so it
	// is added here once and then shared by whatever piles remain.
	energy := w.capacity(a)*a.Temp + w.capacity(b)*b.Temp - prop.rx.Enthalpy*prop.extent

	for _, term := range prop.rx.Reactants {
		p := parcelWithSpecies(term.Species, a, b)
		if p == nil {
			return
		}
		p.Moles -= prop.extent * term.Count
	}

	x := (a.X + b.X) / 2
	y := (a.Y + b.Y) / 2
	born := make([]*Parcel, 0, len(prop.rx.Products))
	for _, term := range prop.rx.Products {
		born = append(born, w.spawn(term.Species, prop.extent*term.Count, x, y, 0))
	}

	parts := make([]*Parcel, 0, 2+len(born))
	if a.Moles > moleEpsilon {
		parts = append(parts, a)
	}
	if b != a && b.Moles > moleEpsilon {
		parts = append(parts, b)
	}
	parts = append(parts, born...)
	w.setTemperature(parts, energy)
}

func (w *World) setTemperature(parts []*Parcel, energy float64) {
	var cap float64
	for _, p := range parts {
		cap += w.capacity(p)
	}
	if cap <= 0 {
		return
	}
	temp := energy / cap
	for _, p := range parts {
		p.Temp = temp
	}
}

func (w *World) phaseChange() {
	// New parcels are appended. Each pass handles parcels born on the previous
	// pass, so water created by a reaction can still freeze or boil this tick.
	start := 0
	for start < len(w.parcels) {
		end := len(w.parcels)
		for i := start; i < end; i++ {
			w.considerPhase(w.parcels[i])
		}
		start = end
	}
}

func (w *World) considerPhase(p *Parcel) {
	if p.Moles <= moleEpsilon {
		return
	}
	switch p.Species {
	case Water:
		if p.Temp < freezeK {
			w.convert(p, Ice, freezeK, -latentFusion)
		} else if p.Temp > boilK {
			w.convert(p, Steam, boilK, latentVapor)
		}
	case Ice:
		if p.Temp > freezeK {
			w.convert(p, Water, freezeK, latentFusion)
		}
	case Steam:
		if p.Temp < boilK {
			w.convert(p, Water, boilK, -latentVapor)
		}
	}
}

// convert moves moles from p into `to` until the pile would sit at target,
// or until p is used up. latentAbsorbed is joules taken in per mole converted.
func (w *World) convert(p *Parcel, to SpeciesID, target, latentAbsorbed float64) {
	from := w.cat.Must(p.Species)
	dest := w.cat.Must(to)
	n := p.Moles
	denom := latentAbsorbed + (dest.HeatCapacity-from.HeatCapacity)*target
	if math.Abs(denom) < 1e-9 || n <= 0 {
		return
	}
	x := from.HeatCapacity * n * (p.Temp - target) / denom
	if x <= 1e-12 {
		return
	}
	if x > n {
		x = n
	}
	energy := from.HeatCapacity*n*p.Temp - latentAbsorbed*x
	p.Moles = n - x
	heatDenom := from.HeatCapacity*p.Moles + dest.HeatCapacity*x
	if heatDenom <= 0 {
		p.Moles = n
		return
	}
	tNew := energy / heatDenom
	if p.Moles <= moleEpsilon {
		p.Moles = 0
	}
	p.Temp = tNew
	w.spawn(to, x, p.X, p.Y, tNew)
}

func (w *World) spawn(species SpeciesID, moles, x, y, temp float64) *Parcel {
	p := &Parcel{
		ID:      w.nextID,
		Species: species,
		Moles:   moles,
		X:       x,
		Y:       y,
		Temp:    temp,
	}
	w.nextID++
	w.parcels = append(w.parcels, p)
	return p
}

// coalesce folds together piles of the same material that sit on top of each
// other, keeping the total moles and the total thermal energy.
func (w *World) coalesce() {
	for i := 0; i < len(w.parcels); i++ {
		a := w.parcels[i]
		if a.Moles <= moleEpsilon {
			continue
		}
		for j := i + 1; j < len(w.parcels); j++ {
			b := w.parcels[j]
			if b.Species != a.Species || b.Moles <= moleEpsilon {
				continue
			}
			if hypot(a.X-b.X, a.Y-b.Y) > mergeDistance {
				continue
			}
			c0 := w.capacity(a)
			c1 := w.capacity(b)
			n := a.Moles + b.Moles
			if c0+c1 > 0 {
				a.Temp = (c0*a.Temp + c1*b.Temp) / (c0 + c1)
			}
			a.X = (a.X*a.Moles + b.X*b.Moles) / n
			a.Y = (a.Y*a.Moles + b.Y*b.Moles) / n
			a.Moles = n
			b.Moles = 0
		}
	}
}

func (w *World) reap() {
	kept := make([]*Parcel, 0, len(w.parcels))
	for _, p := range w.parcels {
		if p.Moles < 0 && p.Moles > -1e-8 {
			p.Moles = 0
		}
		if p.Moles > moleEpsilon {
			kept = append(kept, p)
		}
	}
	w.parcels = kept
}

func (w *World) index() map[int]*Parcel {
	byID := make(map[int]*Parcel, len(w.parcels))
	for _, p := range w.parcels {
		byID[p.ID] = p
	}
	return byID
}

func (w *World) capacity(p *Parcel) float64 {
	return p.Moles * w.cat.Must(p.Species).HeatCapacity
}

func parcelWithSpecies(id SpeciesID, a, b *Parcel) *Parcel {
	if a != nil && a.Species == id {
		return a
	}
	if b != nil && b.Species == id {
		return b
	}
	return nil
}

func distance(a, b *Parcel) float64 {
	return hypot(a.X-b.X, a.Y-b.Y)
}

func hypot(x, y float64) float64 {
	return math.Hypot(x, y)
}

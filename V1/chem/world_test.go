package chem

import "testing"

func TestCatalogBalances(t *testing.T) {
	if err := DefaultCatalog().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSaltDissolvesWhenTouchingWater(t *testing.T) {
	near := newWorld(t)
	mustSpawn(t, near, Salt, 1, 0, 0, 293)
	mustSpawn(t, near, Water, 2, 0.4, 0, 293)
	tick(near, 4, 0.5)

	if got := near.MolesOf(DissolvedSalt); got < 0.9 {
		t.Fatalf("dissolved salt = %.4f, want most of the 1 mol to dissolve", got)
	}
	if near.MolesOf(Water) < 1.9 {
		t.Fatalf("water was consumed, got %.4f mol", near.MolesOf(Water))
	}

	far := newWorld(t)
	mustSpawn(t, far, Salt, 1, 0, 0, 293)
	mustSpawn(t, far, Water, 2, 30, 0, 293)
	tick(far, 4, 0.5)
	if got := far.MolesOf(DissolvedSalt); got > 1e-9 {
		t.Fatalf("far salt dissolved: %.6f mol", got)
	}
}

func TestAcidBaseDependsOnDistanceAndHeat(t *testing.T) {
	cases := []struct {
		name             string
		x                float64
		temp             float64
		ticks            int
		minSalt, maxSalt float64
	}{
		{name: "close and warm", x: 0.3, temp: 450, ticks: 8, minSalt: 0.9, maxSalt: 1.01},
		{name: "far apart", x: 30, temp: 450, ticks: 8, minSalt: 0, maxSalt: 1e-9},
		{name: "close and cold", x: 0.3, temp: 160, ticks: 8, minSalt: 0, maxSalt: 0.01},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := acidBaseWorld(t, tc.x, tc.temp)
			tick(w, tc.ticks, 1)
			got := w.MolesOf(DissolvedSalt)
			if got < tc.minSalt || got > tc.maxSalt {
				t.Fatalf("dissolved salt = %.6f, want between %.6f and %.6f", got, tc.minSalt, tc.maxSalt)
			}
		})
	}
}

func TestHotterPairReactsFurther(t *testing.T) {
	hot := acidBaseWorld(t, 0.3, 450)
	cold := acidBaseWorld(t, 0.3, 300)
	hot.Tick(1)
	cold.Tick(1)
	if hot.MolesOf(DissolvedSalt) <= cold.MolesOf(DissolvedSalt)*5 {
		t.Fatalf("hot %.6f mol should be well ahead of cold %.6f mol", hot.MolesOf(DissolvedSalt), cold.MolesOf(DissolvedSalt))
	}
}

func TestLimitingReactantIsBase(t *testing.T) {
	w := newWorld(t)
	mustSpawn(t, w, Acid, 2, 0, 0, 500)
	mustSpawn(t, w, Base, 0.4, 0.3, 0, 500)
	tick(w, 6, 1)
	if got := w.MolesOf(DissolvedSalt); got < 0.39 || got > 0.41 {
		t.Fatalf("product = %.4f mol, want about 0.4 from the base", got)
	}
	if got := w.MolesOf(Acid); got < 1.55 || got > 1.65 {
		t.Fatalf("leftover acid = %.4f mol, want about 1.6", got)
	}
	if got := w.MolesOf(Base); got > 1e-6 {
		t.Fatalf("base still present: %.6f mol", got)
	}
}

func TestSharedParcelIsNotOverConsumed(t *testing.T) {
	w := newWorld(t)
	mustSpawn(t, w, Base, 1, 0, 0, 800)
	mustSpawn(t, w, Acid, 1, 0.3, 0, 800)
	mustSpawn(t, w, Acid, 1, -0.3, 0, 800)
	before := w.Mass()
	w.Tick(1)

	if got := w.MolesOf(DissolvedSalt); got < 0.98 || got > 1.02 {
		t.Fatalf("product = %.4f mol, one base can neutralize only 1 mol of acid", got)
	}
	if got := w.MolesOf(Base); got > 1e-6 {
		t.Fatalf("base left = %.6f", got)
	}
	if got := w.MolesOf(Acid); got < 0.98 || got > 1.02 {
		t.Fatalf("leftover acid = %.4f mol, want about 1", got)
	}
	assertNoNegative(t, w)
	assertMass(t, before, w.Mass())
}

func TestMassIsConserved(t *testing.T) {
	w := busyWorld(t)
	before := w.Mass()
	for i := 0; i < 10; i++ {
		w.Tick(1)
		assertMass(t, before, w.Mass())
		assertNoNegative(t, w)
	}
}

func TestHeatExchangeConservesEnergyAndMovesTemperatures(t *testing.T) {
	w := newWorld(t)
	mustSpawn(t, w, Water, 10, 0, 0, 300)
	mustSpawn(t, w, Water, 10, 1, 0, 340)
	before := thermalEnergy(w)
	w.Tick(1)
	snaps := w.Snapshot()
	if len(snaps) != 2 {
		t.Fatalf("parcels = %d, heat exchange should not create or destroy water", len(snaps))
	}
	if snaps[0].Temp > snaps[1].Temp {
		t.Fatalf("temperatures crossed: %.2f and %.2f", snaps[0].Temp, snaps[1].Temp)
	}
	if snaps[0].Temp <= 300 || snaps[1].Temp >= 340 {
		t.Fatalf("temperatures did not move together: %.2f and %.2f", snaps[0].Temp, snaps[1].Temp)
	}
	after := thermalEnergy(w)
	if delta := abs(after - before); delta > 1e-4 {
		t.Fatalf("thermal energy changed by %.6f J", delta)
	}
}

func TestWaterFreezesTowardZeroCelsius(t *testing.T) {
	w := newWorld(t)
	mustSpawn(t, w, Water, 10, 0, 0, 263.15)
	before := w.Mass()
	w.Tick(1)
	if w.MolesOf(Ice) < 0.2 {
		t.Fatalf("ice = %.4f mol, expected a visible freeze", w.MolesOf(Ice))
	}
	total := w.MolesOf(Water) + w.MolesOf(Ice)
	if abs(total-10) > 1e-6 {
		t.Fatalf("water+ice = %.6f mol, want 10", total)
	}
	for _, p := range w.Snapshot() {
		if p.Species == Water && abs(p.Temp-freezeK) > 0.05 {
			t.Fatalf("remaining water at %.3f K, want about %.2f K", p.Temp, freezeK)
		}
	}
	assertMass(t, before, w.Mass())
}

func TestRunsMatch(t *testing.T) {
	a := busyWorld(t)
	b := busyWorld(t)
	tick(a, 8, 1)
	tick(b, 8, 1)
	left := a.Snapshot()
	right := b.Snapshot()
	if len(left) != len(right) {
		t.Fatalf("parcel count %d vs %d", len(left), len(right))
	}
	for i := range left {
		if left[i].ID != right[i].ID || left[i].Species != right[i].Species {
			t.Fatalf("parcel %d differs: %+v vs %+v", i, left[i], right[i])
		}
		if abs(left[i].Moles-right[i].Moles) > 1e-9 || abs(left[i].Temp-right[i].Temp) > 1e-9 {
			t.Fatalf("parcel %d state differs: %+v vs %+v", i, left[i], right[i])
		}
	}
}

func TestDissolvingAccountsForEnthalpy(t *testing.T) {
	w := newWorld(t)
	mustSpawn(t, w, Salt, 1, 0, 0, 293)
	mustSpawn(t, w, Water, 2, 0.4, 0, 293)
	before := thermalEnergy(w)
	w.Tick(0.5)
	if w.MolesOf(DissolvedSalt) < 0.99 {
		t.Fatalf("expected the salt to dissolve, got %.4f mol", w.MolesOf(DissolvedSalt))
	}
	// Dissolving this salt absorbs 3800 J per mole.
	want := before - 3800
	if delta := abs(thermalEnergy(w) - want); delta > 1e-2 {
		t.Fatalf("thermal energy = %.3f J, want %.3f J", thermalEnergy(w), want)
	}
}

func TestHotParcelCoolsTowardAir(t *testing.T) {
	w := newWorld(t)
	w.Ambient = 293
	w.AirLoss = 8
	mustSpawn(t, w, Iron, 2, 0, 0, 1100)
	w.Tick(1)
	got := w.Snapshot()[0].Temp
	if got >= 1100 || got <= 293 {
		t.Fatalf("temperature = %.1f K, want it to fall from 1100 toward 293", got)
	}
}

func TestDustCoolsToAirWithoutGoingPastIt(t *testing.T) {
	w := newWorld(t)
	w.Ambient = 293
	w.AirLoss = 100
	mustSpawn(t, w, IronOxide, 0.001, 0, 0, 1500)
	w.Tick(1)
	got := w.Snapshot()[0].Temp
	if got < w.Ambient-1e-6 || got >= 1500 {
		t.Fatalf("temperature = %.2f K, want it to fall from 1500 toward 293 and stop there", got)
	}
}

func TestSpawnRejectsBadInput(t *testing.T) {
	w := NewWorld()
	if err := w.Spawn(SpeciesID("gold"), 1, 0, 0, 300); err == nil {
		t.Fatal("expected unknown species to fail")
	}
	if err := w.Spawn(Water, 0, 0, 0, 300); err == nil {
		t.Fatal("expected zero moles to fail")
	}
	if err := w.Spawn(Water, 1, 0, 0, -5); err == nil {
		t.Fatal("expected negative temperature to fail")
	}
}

func newWorld(t *testing.T) *World {
	t.Helper()
	return NewWorld()
}

func busyWorld(t *testing.T) *World {
	t.Helper()
	w := NewWorld()
	mustSpawn(t, w, Salt, 1, 0, 0, 293)
	mustSpawn(t, w, Water, 2, 0.4, 0, 293)
	mustSpawn(t, w, Acid, 1.2, 12, 0, 430)
	mustSpawn(t, w, Base, 0.5, 12.3, 0, 430)
	mustSpawn(t, w, Iron, 2, 24, 0, 1100)
	mustSpawn(t, w, Oxygen, 3, 24.4, 0, 1100)
	mustSpawn(t, w, Water, 4, 36, 0, 260)
	return w
}

func acidBaseWorld(t *testing.T, x, temp float64) *World {
	t.Helper()
	w := NewWorld()
	mustSpawn(t, w, Acid, 1, 0, 0, temp)
	mustSpawn(t, w, Base, 1, x, 0, temp)
	return w
}

func mustSpawn(t *testing.T, w *World, species SpeciesID, moles, x, y, temp float64) {
	t.Helper()
	if err := w.Spawn(species, moles, x, y, temp); err != nil {
		t.Fatal(err)
	}
}

func tick(w *World, n int, dt float64) {
	for i := 0; i < n; i++ {
		w.Tick(dt)
	}
}

func thermalEnergy(w *World) float64 {
	var energy float64
	for _, p := range w.Snapshot() {
		energy += p.Moles * w.cat.Must(p.Species).HeatCapacity * p.Temp
	}
	return energy
}

func assertNoNegative(t *testing.T, w *World) {
	t.Helper()
	for _, p := range w.Snapshot() {
		if p.Moles < 0 {
			t.Fatalf("%s has negative moles %.6f", p.Name, p.Moles)
		}
	}
}

func assertMass(t *testing.T, before, after float64) {
	t.Helper()
	if abs(after-before) > 1e-4 {
		t.Fatalf("mass %.6f g drifted from %.6f g", after, before)
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

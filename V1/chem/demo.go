package chem

// DemoWorld is the scene shown in the terminal and in the browser.
// Four pairs sit far enough apart that only partners in a pair can react.
func DemoWorld() *World {
	w := NewWorld()
	w.Ambient = 293
	w.AirLoss = 8

	// 293 K is about 20 °C. Salt touching water should dissolve.
	place(w, Salt, 1, 0, 0, 293)
	place(w, Water, 2, 0.4, 0, 293)

	// 450 K is hot, so this acid and base neutralize over a few ticks.
	place(w, Acid, 1, 12, 0, 450)
	place(w, Base, 1, 12.4, 0, 450)

	// Cold iron in oxygen stays iron.
	place(w, Iron, 2, 24, 0, 300)
	place(w, Oxygen, 2, 24.4, 0, 300)

	// 1100 K is red-hot iron. The same pair oxidizes, slowly.
	place(w, Iron, 2, 36, 0, 1100)
	place(w, Oxygen, 2, 36.4, 0, 1100)
	return w
}

func place(w *World, species SpeciesID, moles, x, y, temp float64) {
	if err := w.Spawn(species, moles, x, y, temp); err != nil {
		panic(err)
	}
}

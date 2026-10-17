// Command sandbox places a few chemical parcels and prints every tick.
//
// Run it from the module root:
//
//	go run ./cmd/sandbox
package main

import (
	"fmt"

	"miniverse/chem"
)

func main() {
	world := chem.DemoWorld()

	fmt.Printf("start   mass %.3f g\n", world.Mass())
	printParcels(world)

	for tick := 1; tick <= 6; tick++ {
		world.Tick(1)
		fmt.Printf("\ntick %d   mass %.3f g\n", tick, world.Mass())
		printParcels(world)
	}
}

func printParcels(world *chem.World) {
	for _, p := range world.Snapshot() {
		fmt.Printf("  %-16s %7.3f mol  %7.1f K  at (%4.1f, %4.1f)\n", p.Name, p.Moles, p.Temp, p.X, p.Y)
	}
}

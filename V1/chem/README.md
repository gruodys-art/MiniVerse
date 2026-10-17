# chem

The simulation library. A parcel is one material at a point: species, moles, position, and temperature in kelvin. `World.Tick` advances every parcel by a time step in seconds.

Backend team docs: [architecture](../docs/backend/architecture.md), [layers](../docs/backend/layers.md), [coding standards](../docs/backend/coding-standards.md). This package is the domain layer; adapters live under `cmd/`.

Each tick the world:

1. Shares heat between parcels that touch, and with the surrounding air.
2. Applies reactions from the catalog when the participants are close enough.
3. Changes phase for water, ice, and steam when the temperature crosses a threshold.
4. Merges piles of the same material that sit on top of each other.

Reactions come only from `DefaultCatalog`. Each reaction must balance atoms, and it is either a pair of reactants or one reactant beside a neighbor it needs but does not consume (salt dissolving in water). `Catalog.Validate` checks that shape.

## Entry points

| Function | Role |
| --- | --- |
| `NewWorld` | Empty world with the default catalog |
| `Spawn` | Place a parcel. Moles and temperature must be positive |
| `Tick` | Advance the world by `dt` seconds |
| `Mass` | Total mass in grams |
| `MolesOf` | Moles of one species |
| `Snapshot` | Copy of every parcel, including display names |
| `DemoWorld` | The scene the sandbox and the web server show |

`DemoWorld` places four pairs far enough apart that only partners in a pair can react: salt in water, hot acid and base, cold iron in oxygen, and red-hot iron in oxygen.

## Tests

From the module root:

```text
go test ./chem
```

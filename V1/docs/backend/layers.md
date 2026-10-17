# Backend layers

This is the package and import map of **today’s** layout. Layers are directories and import direction, not a framework.

Module: `miniverse`. Packages: `chem`, `cmd/sandbox`, `cmd/web`.

## Layer map

| Layer | Path | Role |
| --- | --- | --- |
| Domain / simulation | `chem/` | Closed catalog, validation, world, tick, demo scene, snapshot views |
| HTTP delivery | `cmd/web/` | Mutex-serialized API, frame JSON assembly, embed of `dist` |
| Terminal delivery | `cmd/sandbox/` | Print `DemoWorld` and six one-second ticks |

```text
cmd/sandbox ──imports──► chem
cmd/web     ──imports──► chem
chem        ──must not import──► cmd/*, net/http for serving, web/*
```

`chem` may use the standard library (`fmt`, `math`, `sort`, …). It must not import `cmd/sandbox`, `cmd/web`, or the React tree under `web/`.

## Allowed and forbidden imports

| From | May import | Must not import |
| --- | --- | --- |
| `chem` | Go stdlib only (plus future `chem/...` subpackages) | `cmd/*`, `net/http` handlers, embed of UI, anything under `web/` |
| `cmd/web` | `miniverse/chem`, Go stdlib (`net/http`, `encoding/json`, `embed`, …) | Sandbox printing helpers; chemistry formulas living only in the handler |
| `cmd/sandbox` | `miniverse/chem`, Go stdlib (`fmt`, …) | HTTP, JSON frame types owned by web, catalog edits |

Forbidden in practice:

- Putting a new reaction or species only in `cmd/web` or `cmd/sandbox`
- Calling `World.Tick` from React (the client sends `POST /api/tick`)
- `chem` depending on how the frame is served or printed

## What may live where

### `chem` (domain)

- Species, atoms, molar mass, heat capacity
- `DefaultCatalog`, `Catalog.Validate` (shape + atom balance)
- `World`, `Spawn` (rejects unknown species, non-positive moles, non-positive kelvin)
- `Tick` and its private helpers (heat, propose/limit/apply, phase, coalesce, reap)
- `DemoWorld` (four pairs far enough apart that only partners in a pair can react)
- `Snapshot` / `ParcelView` (JSON tags are the parcel half of the frame contract)

### `cmd/web` (HTTP adapter)

- Process-owned `sim` (world + tick counter + mutex)
- Routes: `GET /api/state`, `POST /api/tick`, `POST /api/reset`, static `/` from embedded `dist`
- Frame struct that wraps tick, mass, and `[]chem.ParcelView`
- `go:generate` / embed of the client build — not simulation logic

### `cmd/sandbox` (terminal adapter)

- `main` that constructs `DemoWorld`, prints mass and parcels, ticks six times
- Formatting for humans — not catalog or HTTP

## Ownership for parallel work

| Concern | Primary package | Notes |
| --- | --- | --- |
| Materials and reactions | `chem` (`catalog.go`) | Keep `Validate` green; do not invent chemistry in comments or code |
| Tick order / heat / phase / merge | `chem` (`world.go`) | Physicist/chemist vetoes bind here; do not weaken them in adapters |
| Demo scene layout | `chem` (`demo.go`) | Both sandbox and web show this scene |
| HTTP routes, mutex, embed | `cmd/web` | One world, no session |
| Terminal output | `cmd/sandbox` | Numbers without the browser |
| Browser UI | `web/` (not this backend layer doc) | Consumes frame JSON only |

Two developers can land catalog and HTTP changes in the same week if they stay in their packages and keep the frame contract stable. If a change needs both a new parcel field and a canvas paint, coordinate the JSON shape first (see [coding-standards.md](coding-standards.md)).

## Splitting `chem` later

If files grow large, prefer subpackages under the same module (for example `chem/catalog` imported by `chem` or by a thin `chem` facade) **without** reversing the arrow: adapters still import the simulation library; the library still does not import adapters. Avoid a parallel “domain service” process until the product actually needs more than one world or process.

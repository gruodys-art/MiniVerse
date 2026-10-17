# Go coding standards (MiniVerse backend)

Standards for `chem/`, `cmd/sandbox/`, and `cmd/web/`. Match the voice already in `chem/world.go` and `chem/catalog.go`: plain language, short comments that say why, no invented chemistry.

## Packages and layout

- Module path: `miniverse`.
- Simulation code lives in `chem`. Delivery lives in `cmd/...` as `package main`.
- Prefer small, obvious names (`World`, `Spawn`, `Tick`, `frame`, `sim`) over abstractions that do not exist in the product.
- Keep import direction one-way: adapters → `chem`. See [layers.md](layers.md).

## Naming

- Exported identifiers are for callers (`DemoWorld`, `DefaultCatalog`, `ParcelView`).
- Unexported helpers stay next to the algorithm they serve (`shareHeat`, `propose`, `limitExtents`).
- Species and reaction IDs are stable strings (`"water"`, `"dissolve_salt"`). Do not rename lightly; the client and demos depend on them.
- JSON field names are lowercase (`json:"temp"`). Do not camelCase the wire format without updating TypeScript.

## Errors and validation

- `Spawn` returns an error for unknown species, non-positive moles, or non-positive kelvin. Callers that build the demo may `panic` on that error only when a programmer mistake made the scene invalid.
- `Catalog.Validate` returns errors for bad reaction shape, unbalanced atoms, missing molar mass, or unknown elements. `NewWorld` panics if the default catalog fails validation — that is a programming mistake, not a runtime input.
- `Tick` with non-positive `dt` is a no-op (no error).
- Prefer `fmt.Errorf` with a clear message. Do not wrap HTTP status inventiveness into `chem`.

## Comments

- Package and exported comments should read like the existing READMEs: what the thing is, in everyday words.
- Explain tick order, energy bookkeeping, and extent limiting when the code is dense — not textbook chemistry the catalog does not model.
- Do not claim reactions or materials that are not in `DefaultCatalog`.

## JSON and the frame contract

Wire shape for `cmd/web`:

```json
{
  "tick": 0,
  "mass": 0,
  "parcels": [
    {
      "id": 1,
      "name": "water",
      "species": "water",
      "moles": 1,
      "x": 0,
      "y": 0,
      "temp": 293
    }
  ]
}
```

- Parcel fields are defined by `chem.ParcelView` tags; the frame wrapper lives in `cmd/web`.
- The TypeScript copy is `Frame` / `Parcel` in `web/src/types.ts`.
- Any change to field names, units, or nesting is a **contract change**: update Go, TypeScript, `cmd/web/README.md`, and [architecture.md](architecture.md) in the same effort. Prefer not to change the contract unless the product needs it.

## Tests

- Library: `go test ./chem` from the module root.
- HTTP: `go test ./cmd/web` after the client is built into `cmd/web/dist` (`npm --prefix web run build` or `go generate ./cmd/web`).
- Full module: `go test ./...` (same embed requirement for `cmd/web`).
- Prefer tests that lock catalog balance, spawn rejection, tick mass behavior, and API frame shape over speculative mocks.
- Do not add tests that require inventing chemistry outside the catalog.

## When to update READMEs

Update the package README when you change:

- Entry points or tick behavior users rely on (`chem/README.md`)
- API routes or frame fields (`cmd/web/README.md`)
- How to run the sandbox (`cmd/sandbox/README.md`)
- Top-level run/test story (`README.md`)

Also keep [architecture.md](architecture.md) and [layers.md](layers.md) truthful if runtime model or package boundaries change. Doc-only backend guides live under `docs/backend/`.

## Review checklist (boundaries)

Before merging Go changes, confirm:

1. **Import direction** — `chem` does not import `cmd/*` or `web/*`.
2. **Layer content** — catalog/reactions only in `chem`; HTTP/JSON serving only in `cmd/web`; printing only in `cmd/sandbox`.
3. **Frame contract** — if JSON changed, TypeScript and READMEs match; if unchanged, say so in the change notes.
4. **Vetoes** — physicist or chemist vetoes are not weakened in adapters or “temporary” tick shortcuts.
5. **Demo coherence** — sandbox and web still show `DemoWorld` unless the product deliberately diverges.
6. **Tests** — `go test ./chem` (and `./cmd/web` if HTTP touched) pass with embed rules respected.
7. **Comments** — plain language; no chemistry that is not in the catalog.

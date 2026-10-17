# Backend architecture

MiniVerse’s simulation and its Go entry points share one picture: a single process holds one world, and adapters only read or command that world. This document is for the Go backend team. It describes what exists today, not a future service mesh.

## Goals

- Keep the chemical world coherent: one catalog, one tick order, one snapshot shape.
- Let many developers change catalog, tick physics, HTTP, or sandbox printing without stepping on each other’s packages.
- Preserve a one-way dependency: adapters depend on `chem`; `chem` never depends on HTTP, UI, or printing.

## Runtime model (today)

| Fact | Today |
| --- | --- |
| Processes | One Go process per demo (`cmd/sandbox` or `cmd/web`) |
| Worlds | One `chem.World` in that process |
| Sessions / auth | None |
| Clock | Callers choose `dt`; the web API advances one second per `POST /api/tick` |
| Concurrency | `cmd/web` serializes API handlers with a mutex; `chem` itself is not concurrent |

There is no multi-world server, no worker pool, and no shared database. The browser paints what the server last returned; it does not advance the simulation on its own.

## How the pieces relate

```text
  ┌─────────────────────────────────────────┐
  │  cmd/sandbox     cmd/web                │  delivery adapters
  │  (print ticks)   (HTTP + embed dist)    │
  └───────────────────┬─────────────────────┘
                      │ import miniverse/chem
                      ▼
  ┌─────────────────────────────────────────┐
  │  chem                                   │  domain / simulation library
  │  catalog, World.Tick, DemoWorld,        │
  │  Snapshot / ParcelView                  │
  └─────────────────────────────────────────┘
```

- **`chem`** owns materials, reactions, parcels, heat, phase, and coalescence. It is a library, not a server.
- **`cmd/sandbox`** builds `DemoWorld`, prints mass and parcels, then ticks six times. Delivery is stdout.
- **`cmd/web`** holds `DemoWorld` behind a mutex, serves JSON frames, and embeds the built React client from `dist`. Delivery is HTTP.

Both adapters call the same library. Neither reimplements chemistry.

## Request and tick path (`cmd/web`)

1. HTTP request hits `GET /api/state`, `POST /api/tick`, or `POST /api/reset`.
2. Handler takes the process mutex (one world; requests do not interleave).
3. On tick: `World.Tick(1)` then increment the server tick counter. On reset: replace with `chem.DemoWorld()` and zero the counter. On state: no mutation.
4. Build a frame: `tick`, `mass` (`World.Mass()`), `parcels` (`World.Snapshot()`).
5. Encode JSON and return. Parcel fields come from `chem.ParcelView` JSON tags.

`World.Tick` itself (when `dt > 0`) runs in fixed order: share heat, shed heat to air, propose reactions, limit extents so piles are not double-spent, apply reactions, phase change, coalesce nearby same-species piles, reap empty piles. If `dt` is not positive, `Tick` returns immediately.

## Frame contract

The HTTP frame is `tick`, `mass` (grams), and `parcels`. Each parcel is `id`, `name`, `species`, `moles`, `x`, `y`, and `temp` (kelvin). The TypeScript mirror is `Frame` in `web/src/types.ts`. Changing that JSON is a cross-team contract change: update Go tags, TypeScript types, package READMEs, and this architecture note together.

## Non-goals (intentionally out of scope)

- Microservices, message buses, or a separate “API gateway”
- Multiple concurrent worlds, user sessions, or authentication
- Inventing reactions outside `DefaultCatalog`
- Advancing the world inside the React client
- Hexagonal ports/adapters layers that do not match the three packages above

## Growing without breaking the picture

When `chem` becomes large, split **inside** the module under packages that still export a clear library API, and keep `cmd/*` as thin adapters. New delivery (CLI flags, a second HTTP surface) belongs under `cmd/`, importing `chem`, never the reverse. Do not introduce a fake middle tier “for architecture” if the only callers are still sandbox and web.

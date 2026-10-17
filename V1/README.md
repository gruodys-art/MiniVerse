# MiniVerse

MiniVerse is a small chemical simulation. Materials sit on a plane as parcels. Each tick, parcels that are close enough share heat and may react. The reaction list is a closed catalog: the simulation does not invent new chemistry.

The world lives in Go. The terminal sandbox and the browser both watch that same demo scene. The browser paints the latest server snapshot and sends commands back. It does not advance the simulation on its own.

## Projects

| Project | What it is |
| --- | --- |
| [chem](chem/README.md) | Simulation library: parcels, catalog, ticks |
| [cmd/sandbox](cmd/sandbox/README.md) | Prints the demo scene, tick by tick |
| [cmd/web](cmd/web/README.md) | Serves the world over HTTP and embeds the built page |
| [web](web/README.md) | React shell and canvas. Reads the server frame |

## Backend docs

For Go layout, import boundaries, and coding standards (team-facing):

- [Architecture](docs/backend/architecture.md) — one process, one world, request/tick path
- [Layers](docs/backend/layers.md) — `chem` vs `cmd/*` ownership and imports
- [Coding standards](docs/backend/coding-standards.md) — Go style, frame contract, review checklist


## Run the demo

From this directory, with Go and Node installed:

```text
go run ./cmd/sandbox
```

```text
npm --prefix web install
npm --prefix web run build
go run ./cmd/web
```

Then open http://127.0.0.1:8080.

While changing the page, leave the Go server running and start the client dev server:

```text
npm --prefix web run dev
```

Vite serves the page at http://127.0.0.1:5173 and proxies `/api` to the Go process.

## Tests

```text
go test ./...
```

`cmd/web` embeds `cmd/web/dist`, so build the client before `go test` on a fresh checkout. `go generate ./cmd/web` runs that build.

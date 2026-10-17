# web server

Serves one `chem.DemoWorld` and the built browser client. The process holds the world. Every API call returns the full frame after the command.

Backend team docs: [architecture](../../docs/backend/architecture.md), [layers](../../docs/backend/layers.md), [coding standards](../../docs/backend/coding-standards.md). This command is the HTTP adapter around `chem`; it does not own the catalog or tick order.

The client source is [web](../../web/README.md). This command embeds the build output in `dist`. Build that before running or testing:

```text
npm --prefix web install
npm --prefix web run build
go run ./cmd/web
```

Or, from this directory, `go generate` runs the client build. Then open http://127.0.0.1:8080.

For UI work, keep this server running and use `npm --prefix web run dev`. Vite proxies `/api` here.

## API

| Request | Effect |
| --- | --- |
| `GET /api/state` | Current frame |
| `POST /api/tick` | Advance one second, then return the frame |
| `POST /api/reset` | Replace the world with a new demo scene |
| `GET /` | Built client from `dist` |

The frame JSON is `tick`, `mass` in grams, and `parcels`. Each parcel has `id`, `name`, `species`, `moles`, `x`, `y`, and `temp` in kelvin. The TypeScript copy of that shape is `Frame` in `web/src/types.ts`.

One server, one world. Requests are serialized. There is no session yet.

## Tests

From the module root, after the client has been built:

```text
go test ./cmd/web
```

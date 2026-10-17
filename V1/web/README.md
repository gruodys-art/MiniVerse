# web client

React shell for the MiniVerse page. The Go server in [cmd/web](../cmd/web/README.md) owns the world. This app fetches a frame and draws it.

World state is the latest frame (`useWorld`): tick, mass, and parcels. View state (`useView`) is browser-only. Today that is play and pause, which decides when to call `POST /api/tick`.

The canvas is a separate core. `draw.ts` paints one frame and does not fetch or store anything. `Stage` calls it. Header and the parcel table are shell components. A chart or data-grid library belongs in `src/components` and should read `useWorld().frame`.

Species colors live in `palette.ts`. They are presentation, not part of the snapshot.

## Scripts

From this directory:

```text
npm install
npm run dev
npm run build
```

`dev` serves http://127.0.0.1:5173 and proxies `/api` to http://127.0.0.1:8080, so `go run ./cmd/web` needs to be running.

`build` typechecks, then writes the bundle to `../cmd/web/dist`. The Go server embeds that directory. `cmd/web/dist` is generated and is not source.

## Layout

| Path | Role |
| --- | --- |
| `src/api.ts` | `GET /api/state`, `POST /api/tick`, `POST /api/reset` |
| `src/types.ts` | `Frame` and `Parcel`, matching the Go JSON |
| `src/world.tsx` | World frame and view state |
| `src/draw.ts` | Canvas painter |
| `src/components` | Shell: header, stage, table |

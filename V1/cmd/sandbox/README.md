# sandbox

A terminal view of `chem.DemoWorld`. It prints the starting mass and parcels, then six ticks. One tick is one second.

Backend team docs: [architecture](../../docs/backend/architecture.md), [layers](../../docs/backend/layers.md), [coding standards](../../docs/backend/coding-standards.md). This command is a terminal adapter around `chem`; printing only lives here.

From the module root:

```text
go run ./cmd/sandbox
```

Each line is one parcel: name, moles, temperature in kelvin, and position. Use this when changing the simulation and you want the numbers without the browser.

# inflow-inspector-api

**The reference backend for the Inflowenger platform** — and, not incidentally, the working proof that wiring a product into Inflowenger is small.

`inflow-inspector-api` is a Fiber-based Go API that started as the backend for Inflowenger's own developer/admin panel. It's the first real consumer of [`inflow-fusion`](https://github.com/Inflowenger/inflow-fusion), and it happens to implement *everything* a backend needs to fully participate in the platform: CRUD for flows, CRUD for execution context, an extension/plugin catalog, process start/stop, and a realtime log feed for a frontend panel. If you're building your own product on Inflowenger, this repo is the concrete example to copy from.

> **Status:** early-stage / pre-1.0, tracking `inflow-fusion` (see [go.mod](go.mod)).

## The one-sentence version

Inflowenger's execution engine doesn't store anything — it asks your backend for flows and context over NATS, and your backend answers. `inflow-fusion` handles that NATS plumbing; **all you have to build is CRUD for two record types (flows, context) plus a thin adapter between your storage and three interface methods.** That's it. This repo is that "it" — see [docs/wiring.md](docs/wiring.md) for exactly how little code is involved.

## What's in here

| Concern | How it's implemented |
|---|---|
| Flow storage | `models.FlowRecord` (a title + a `view_flow` graph) in an embedded [BadgerDB](https://github.com/dgraph-io/badger) store |
| Context storage | `models.ContextRecord` (a title + an opaque JSON-string `context` blob) in the same store |
| Extension/plugin catalog | `models.ExtensionRecord` — metadata a frontend palette uses to render an extrinsic/plugin node, plus credential issuance for plugins |
| Inflowenger wiring | `inflow/wire.go`, `inflow/port.go`, `inflow/compiler.go` — implements `inflow-fusion`'s `IInflowService` and the Vue Flow → node compiler hook |
| HTTP API | Fiber v3, JSON via `sonic`, single shared HS256 JWT for auth |
| Realtime | Socket.IO endpoint that mirrors the platform's NATS event log to connected frontend clients |
| Infra passthrough | A generic reverse proxy at `/infra/*` so a frontend panel can talk to Inflowenger infra's own REST API without a separate CORS/auth setup |

Full endpoint-by-endpoint reference: [docs/api.md](docs/api.md). A ready-to-import [Insomnia](https://insomnia.rest) collection with real example bodies for every endpoint below ships at [`Insomnia_2026-07-12.yaml`](Insomnia_2026-07-12.yaml).

## Running it

```bash
go get ./...
go run .
```

Configuration is read from environment variables (or a `.env` file — see `env/vars.go`):

| Env var | Default | Purpose |
|---|---|---|
| `PORT` | `8025` | HTTP port this API listens on |
| `DB_STORE_PATH` | `db` | BadgerDB data directory |
| `INFLOW_INFRA_API` | — | Base URL of Inflowenger infra's REST API |
| `INFLOW_INFRA_JWT_SECRET` | — | HMAC secret shared with infra, **and** used to sign/verify the bearer JWT this API itself requires on every request (see [docs/api.md#authentication](docs/api.md#authentication)) |

## Building your own Inflowenger backend

If you're evaluating Inflowenger for your own product rather than extending this panel, here's the whole recipe, in the order you'd actually build it:

1. **Storage for two record types.** A `Flow` (a title + a graph — nodes and edges, however your editor exports them) and a `Context` (a title + an opaque JSON blob representing whatever a running process needs to read/write). See `models/flow.go`, `models/context.go`, `repository/flow.go`, `repository/context.go` here for the minimal shape — this repo's versions are genuinely all there is.
2. **A CRUD API over that storage**, so your frontend panel can create/list/edit/delete flows and context documents. `api/flow`, `api/context` here are a complete, working example.
3. **`inflow.InitBackend` + `IInflowService`.** Three methods — `RetrieveFlow`, `RetrieveContext`, `UpdateContext` — each just a few lines translating a NATS message into a call against the storage from step 1. See `inflow/wire.go`.
4. **A compiler hook**, mapping your frontend's node types to `inflow-fusion` node types. If your frontend is Vue Flow or React Flow, you can lean directly on `inflow-fusion`'s shipped compiler — see `inflow/compiler.go` here for a complete `NodeBuilder` hook, and [inflow-fusion's compiler docs](https://github.com/Inflowenger/inflow-fusion/blob/main/docs/compilers/README.md) for the general pattern.
5. **A visual editor frontend** (Vue Flow / React Flow) that calls your CRUD API and exports the graph shape your compiler hook expects.

Everything past that — NATS credentials, engine discovery/round-robin, running/stopping processes, plugin isolation — is handled for you by `inflow-fusion`. See its [architecture](https://github.com/Inflowenger/inflow-fusion/blob/main/docs/architecture.md) and [protocol](https://github.com/Inflowenger/inflow-fusion/blob/main/docs/protocol.md) docs for how the pieces underneath this repo actually talk to each other.

## Documentation

- [docs/api.md](docs/api.md) — every HTTP/WS endpoint this backend exposes, grouped and with example bodies
- [docs/wiring.md](docs/wiring.md) — a closer look at `inflow/wire.go` and `inflow/compiler.go`: the actual glue between this repo's storage and the Inflowenger platform
- [`inflow-fusion` docs](https://github.com/Inflowenger/inflow-fusion/tree/main/docs) — architecture, protocol, node types, and compilers, shared across any backend built this way

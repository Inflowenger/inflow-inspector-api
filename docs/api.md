# API reference

All routes are mounted at the API's root (`http://localhost:8025` by default — see `env.GetApiPort`). Every example below is drawn from the shipped [Insomnia collection](../Insomnia_2026-07-12.yaml) (`baseUrl` = `http://localhost:8025`, `InfraProxy` = `{{ baseUrl }}/infra`) — import that file for a ready-made client instead of retyping these.

## Authentication

Every route below (except the raw Socket.IO upgrade, which checks the token itself) is behind:

```
Authorization: Bearer <JWT>
```

The JWT must be signed `HS256` with `INFLOW_INFRA_JWT_SECRET` (`jwtware` only validates the signature/algorithm, not specific claims — the Insomnia collection's token has payload `{"admin": true}`, but any HS256 token signed with the right secret is accepted). This is the **same secret** used to talk to Inflowenger infra — one shared secret authenticates both your frontend-to-backend calls and this backend's backend-to-infra calls.

## Flows — `/flow`

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/flow` | Create (no `id`) or update (with `id`) a flow |
| `GET` | `/flow?per_page=&cursor=` | List flows, newest first |
| `GET` | `/flow/id/:flowId` | Get one flow. `:flowId` may be the bare sequence number (`22`) or the full key (`flow:22`) |
| `DELETE` | `/flow/id/:flowId` | Delete a flow |

Body shape (`models.FlowRecord`):

```json
{
  "id": "flow:12",
  "title": "basic workflow-template-2",
  "view_flow": {
    "nodes": [ /* Vue Flow / React Flow node objects — id, type, data, position, handleBounds, ... */ ],
    "edges": [ /* source, target, sourceHandle, targetHandle, data.tags, ... */ ]
  }
}
```

`view_flow` is decoded straight into `inflow-fusion`'s `compilers/vueFlow.VueFlow` type — see [inflow-fusion's compiler docs](https://github.com/Inflowenger/inflow-fusion/blob/main/docs/compilers/vueflow.md) for the exact node/edge shape, and `inflow/compiler.go` in this repo for which `node.data.type` strings this backend's own compiler hook understands (`startNode`, `code`, `contract`, `extrinsic`, `goto`, `pluginNative`, `void`, plus the custom `my_a_ext` extrinsic-by-reference type). A graph needs exactly one `startNode`-typed node — `/ps` and `/ps/compile` both fail otherwise.

## Context — `/context`

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/context` | Create (no `id`) or update (with `id`) a context document |
| `GET` | `/context?per_page=&cursor=` | List context documents, newest first |
| `GET` | `/context/id/:contextId` | Get one. `:contextId` may be bare (`25`) or full (`ctx:25`) |
| `DELETE` | `/context/id/:contextId` | Delete |

Body shape (`models.ContextRecord`):

```json
{
  "id": "ctx:25",
  "title": "context-json-v2",
  "context": "{\"userId\":1023,\"username\":\"coder_99\",\"isActive\":true,\"role\":\"Administrator\"}"
}
```

`context` is a **string that must itself be valid JSON** (the API rejects it otherwise) — it's kept opaque on purpose. This exact string is what gets handed to the engine verbatim as `ContextDoc.Data` when a flow node reads or writes it (see `inflow/wire.go`'s `RetrieveContext`/`UpdateContext` and [inflow-fusion's protocol docs](https://github.com/Inflowenger/inflow-fusion/blob/main/docs/protocol.md)) — your backend never needs to understand its schema, only round-trip it.

## Process — `/ps`

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/ps` | Start a process: look up the flow's start node, hand it to `inflow-fusion` to run on a registered engine |
| `POST` | `/ps/compile` | Compile the flow and return the resulting node map **without** running it |
| `POST` | `/ps/stop/:pid` | Stop a running process |

`POST /ps` and `POST /ps/compile` body (`models.ProcessRequestInput`):

```json
{
  "flowId": "flow:22",
  "contextId": "ctx:25"
}
```

`POST /ps` response: `{"pid": "...", "selected_resource": "http://..."}` — the engine instance chosen by `inflow-fusion`'s round-robin pool actually started running the flow.

`POST /ps/compile` response: `{"selected_resource": ..., "process_req": ..., "compiled": { /* map[nodeId]models.Node */ }}` — useful for validating a graph and inspecting exactly what would be sent, before committing to a real run. Note: as shipped, this handler pins its resource lookup to a hardcoded local hostname (`inflow/process.go`) — treat `selected_resource`/`process_req` from this endpoint as illustrative, not necessarily where `/ps` would actually dispatch to.

`POST /ps/stop/:pid` body (`models.StopRequest`):

```json
{ "resource": "http://<engine-host>:9001" }
```

## Extensions & plugins — `/extension`

"Extensions" are catalog metadata describing an extrinsic or plugin node for a frontend palette (icon, form schema, and — for extrinsics — which registered subject it binds to). This is distinct from the actual runtime subject handlers registered in Go code via `svcHandler.ImplHandlerOnSubject` (see `inflow/wire.go`'s `LoadSvcNodehandlers`).

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/extension` | Create (no `id`) or update (with `id`) an extension record |
| `GET` | `/extension?per_page=&cursor=` | List extension records |
| `GET` | `/extension/id/:extId` | Get one |
| `DELETE` | `/extension/id/:extId` | Delete |
| `GET` | `/extension/extrinsics` | List the extrinsic subjects actually registered at runtime in this process (`svcHandler.GetAllSvcs()`) |
| `POST` | `/extension/plugin/cred` | Issue NATS credentials (and a ready-to-use `.env` block) for a plugin instance |

`POST /extension` body (`models.ExtensionRecord`):

```json
{
  "name": "AddRisks",
  "description": "",
  "type": "extrinsic",
  "icon": { "class": "heroicons", "name": "puzzle-piece", "meta": {} },
  "params": { "schema": {}, "ui": {} },
  "bindTo": { "topic_key": "exports_db", "values": { "TABLE_NAME": "risks" } }
}
```

`bindTo.topic_key` is resolved at compile time via `svcHandler.GetSvc(topic_key)` and `bindTo.values` fills that subject's `{placeholder}` patterns — see the `NODE_MY_A` case in `inflow/compiler.go`.

`POST /extension/plugin/cred` body (`models.CredRequest`):

```json
{
  "name": "mapper-plugin",
  "pluginId": "aa-bbb-ccc-dddd",
  "access": "strict",
  "spaceId": ""
}
```

- `access: "strict"` scopes the credential to subjects owned by `pluginId` only (`InfraSpaces.PluginCredentialStrictPermission`).
- `access: "multi"` issues an unrestricted credential on the account (`InfraSpaces.PluginCredentialOpenPermission`) — useful for a plugin that legitimately needs to talk on more than its own namespace.
- `spaceId` picks which infra account to mint the credential under; omit it to use the builtin plugins account.

Response: `{"cred": "<base64 nats creds>", "env": "INFRA_CRED={cred}\nINFRA_URL={url}\nPLUGIN_ID={pluginId}"}` — the `env` string is meant to be dropped directly into a plugin process's environment.

## Infra proxy — `/infra/*`

Everything under `/infra/` is forwarded as-is (path + query string) to `INFLOW_INFRA_API`, behind the same bearer auth as the rest of this API. This lets a frontend panel talk to Inflowenger infra's own REST API through this backend, without a separate CORS/auth setup for infra directly.

| Example | Forwards to |
|---|---|
| `POST /infra/register/:name` | Register a new engine instance with infra |
| `GET /infra/inflow/resource?per_page=&cursor=` | List registered engine instances |
| `GET /infra/inflow/trace?per_page=&cursor=` | Trace recently finished processes |
| `GET /infra/account/list?per_page=&cursor=` | List infra accounts ("spaces") |

Whatever infra exposes under these paths is available here — this proxy doesn't restrict which infra routes are reachable.

## Realtime — WebSocket / Socket.IO

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/ws/:id` | Socket.IO upgrade; `:id` is a client-chosen session id |
| `POST` | `/sendto` | Push a message to a specific connected session |

`/ws/:id` requires `Authorization: Bearer <JWT>` (as a header or `?Authorization=` query param, since browser WebSocket clients can't set arbitrary headers) — validated inline in `wsAuthHandler` rather than through the shared middleware. Once connected, every session automatically receives everything published on the `inflow.event.log` NATS subject (i.e. live log events from the platform), broadcast to all connected clients.

`POST /sendto` body:

```json
{ "sessId": "devpanel", "message": "alooowwwwwwwwww" }
```

Looks up the session id registered at connect time and emits `message` to that one client only (`SendToSession`).

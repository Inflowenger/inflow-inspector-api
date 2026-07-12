# Wiring this backend to Inflowenger

This is the part that matters if you're deciding whether to build on Inflowenger: here is *all* of the code that connects this repo's storage to the platform. There's no hidden layer — what's below is the complete wiring, not a simplified excerpt.

## The three files

| File | Role |
|---|---|
| `inflow/wire.go` | Implements `inflow-fusion`'s `IInflowService` — answers the engine's three questions by calling into `repository/` |
| `inflow/port.go` | Calls `inflow-fusion`'s `InitBackend` on startup, and registers this backend's one extrinsic subject handler |
| `inflow/compiler.go` | Maps this product's frontend node palette to `inflow-fusion` node types, via the shipped Vue Flow/React Flow compiler |

## Step 1 — implement `IInflowService`

Three methods, each just translating a NATS message into a call against the CRUD storage from `repository/`:

```go
// inflow/wire.go
type InflowWire struct{}

func (isvc *InflowWire) RetrieveFlow(msg *nats.Msg) {
	flowId := strings.Split(msg.Subject, ".")[len(strings.Split(msg.Subject, "."))-1]
	rec, err := repository.GetFlowById(flowId)
	// ... compile rec.ViewFlow into inflowModels.Flow (see step 3), respond with it
}

func (isvc *InflowWire) RetrieveContext(msg *nats.Msg) {
	ctxId := strings.Split(msg.Subject, ".")[len(strings.Split(msg.Subject, "."))-1]
	rec := repository.GetContextById(ctxId)
	// ... respond with inflowModels.ContextDoc{Header: rec.Header, Data: rec.Context}
}

func (isvc *InflowWire) UpdateContext(msg *nats.Msg) {
	var newCtxData inflowModels.ContextDoc
	sonic.Unmarshal(msg.Data, &newCtxData)
	// ... load the existing ContextRecord, overwrite .Context/.Header, repository.UpsertContext(...)
}
```

Nothing here is Inflowenger-specific logic — it's ID parsing plus calls to `repository.GetFlowById` / `GetContextById` / `UpsertContext`, the same functions the HTTP CRUD handlers in `api/flow` and `api/context` use. **The engine and your REST API read/write the exact same records.**

## Step 2 — start the connection

```go
// inflow/port.go
func InitInflowConnection() error {
	return fuse.InitBackend(
		fuse.WithImplementedBackendBy(&InflowWire{}),
		fuse.WithJwtSecretKey(env.GetInfraJWTSecret()),
		fuse.WithInfraApi(env.GetInfraApiUrl()),
	)
}
```

Called once from `main()` (`app.go`) before the HTTP server starts listening. This is the entire integration handshake — `inflow-fusion` takes it from here (fetches NATS credentials, subscribes to the three request subjects, loads the engine resource pool). See [inflow-fusion's architecture doc](https://github.com/Inflowenger/inflow-fusion/blob/main/docs/architecture.md) for what happens on the other side of this call.

The same file also registers this product's one custom extrinsic handler:

```go
func LoadSvcNodehandlers() error {
	return svcHandler.ImplHandlerOnSubject("exports_db", svcHandler.SvcTopic("svc.add.issue.{TABLE_NAME}"),
		func(header nats.Header, data []byte) ([]byte, error) {
			table := strings.Split(header.Get("recv_subject"), ".")[3]
			// ... persist `data` into `table`, however this product's domain wants to
			return []byte(fmt.Sprintf(`{"status":"saved successfully on %s table"}`, table)), nil
		})
}
```

This is what an `ExtensionRecord{Type: "extrinsic", BindTo: {TopicKey: "exports_db", Values: {"TABLE_NAME": "risks"}}}` resolves to at compile time — a flow node that, at runtime, calls this handler with the table name substituted in. Nothing about this handler is special-cased by `inflow-fusion`; it's a plain `svcHandler.ImplHandlerOnSubject` registration, exactly as documented in [inflow-fusion's protocol doc](https://github.com/Inflowenger/inflow-fusion/blob/main/docs/protocol.md#nats-your-backends-own-extrinsic-services).

## Step 3 — the compiler hook

`view_flow` on a `FlowRecord` is stored as raw Vue Flow/React Flow JSON (nodes + edges, positions, handles — whatever the graph editor exported). It is **not** directly executable; `inflow/compiler.go`'s `NodeBuilder` is the hook handed to `inflow-fusion`'s shipped compiler that turns it into real node map:

```go
func FLowCompiler(f models.FlowRecord) (string, map[string]*inflowModels.Node, error) {
	startNodeId, err := GetStartNodeId(f)          // finds the node with data.type == "startNode"
	cmpr := compiler.NewVueFlowCompiler(compiler.WithEachNodeFunc(NodeBuilder))
	return startNodeId, ..., cmpr.Compile(startNodeId, f.ViewFlow)
}
```

`NodeBuilder` switches on this product's frontend node type strings (`code`, `contract`, `extrinsic`, `goto`, `pluginNative`, `void`, plus the custom `my_a_ext`) and, for each, constructs the matching builder from `inflow-fusion`'s `nodes` package (`nodes.NewJsNode`, `nodes.NewOpaRuleLogicNode`, `nodes.NewExtrinsicSvcNode`, ...). This is the one piece of the wiring that's genuinely product-specific — your frontend's palette will have different node types and different `data` fields, so your hook will look different. See [inflow-fusion's compiler docs](https://github.com/Inflowenger/inflow-fusion/blob/main/docs/compilers/vueflow.md) for the general contract, and this file for a complete worked example.

`FLowCompiler` is called from two places: `inflow/wire.go`'s `RetrieveFlow` (compiling on-the-fly every time the engine asks for a flow), and `api/flow/process.go`'s `/ps/compile` endpoint (compiling on-demand so a frontend can validate a graph before running it).

## That's the whole integration

Add up `wire.go` + the `InitBackend` call in `port.go` + one `NodeBuilder` switch statement, and that's every line of code that's aware Inflowenger exists. Everything else in this repo — the Fiber routes, the BadgerDB storage, the Socket.IO log stream — is ordinary backend plumbing you'd build the same way regardless of which workflow engine sat behind it.

If you're wiring up a different product: keep your own storage and CRUD API (steps analogous to `repository/` and `api/flow`, `api/context`), copy the shape of `wire.go` and `port.go` almost verbatim (they don't do anything product-specific), and write your own `NodeBuilder`-equivalent for step 3 matching your frontend's actual node palette.

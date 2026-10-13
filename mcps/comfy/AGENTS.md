# comfy

Generic MCP server for comfy-headless — one MCP tool per workflow in the shared store. See `README.md` for usage; this file is the maintenance guide.

## Concept

Nothing about any workflow is hardcoded here. The server is a thin shell around three things:

1. `internal/store` — reads `COMFY_WORKFLOWS_DIR` (`*.workflow.json`) and derives, per file: tool name (filename slug), token list + types, and LoadImage image inputs.
2. `internal/server` — turns each `store.Workflow` into an MCP tool (`mcp.Server.AddTool` with a generated JSON Schema) and runs calls through the headless client.
3. `internal/headless` — the shared comfy-headless client (fill `{{tokens}}` → POST `/infer-async` → poll `/status` → decode outputs).

Adding/removing a workflow = add/remove a file in the store. No code change.

## Invariants (enforced at startup, tested)

- Unparsable workflow, tool-name collision (`a-b` vs `a_b`), or a workflow with zero tokens → **startup fails**. A half-registered server is worse than a dead one.
- Image parameter names never collide with token names (`_image` suffix) or each other (`_2`, `_3`).
- Token order is **sorted** for determinism (Go map walks are randomized).
- Types come from the shared token standard's numeric set — mirror of `modal-ui/src/lib/run.ts` (`NUMERIC_TOKENS` + `NUMERIC_PATTERNS`). If the standard gains a numeric token, update BOTH `run.ts` and `store.go`'s `numericTokens`.
- `allow-empty` tokens (`negative_prompt`, `clothes`) are optional in the schema and default to `""` in `collectValues`; all other tokens are `required` and fail loudly when absent.
- Handler failures are returned as **tool results with `IsError:true`** (text content), never as JSON-RPC errors — the raw `mcp.ToolHandler` path returns a Go error as an RPC error which clients surface worse (learned while testing; see `server.go` `handlerFor`).

## Testing pattern (fully hermetic)

`internal/server/server_test.go` is the reference E2E: fixture store (temp dir, two workflows) → real `store.Load` → `Build` + `mcp.NewInMemoryTransports()` → mock comfy-headless (`/infer-async` + `/status` on `httptest`) with a request capture. Assertions: generated schemas, submitted JSON (values landed, `{{` never shipped), base64 uploads equal the input file bytes, outputs written + inlined, error paths.

For a live check, write a throwaway `cmd/smoketest` that spawns the built binary over `mcp.CommandTransport` and calls a real tool (it builds argument sets from the tool's own schema) — then delete it.

## Ops

- Binary: `go build -o comfy .` → `~/agents-skills/mcps/comfy/comfy`.
- Registered in Hermes `mcp_servers` as `comfy`; requires a raised per-server `timeout` (1800s) — GPU renders exceed the 120s default.
- RESTART the server (or `/reload-mcp`) after adding workflows to the store: the tool list is built once at startup by design.

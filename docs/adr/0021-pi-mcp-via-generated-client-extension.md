# ADR 0021: PI MCP via a gyroscope-generated client extension (official SDK)

- **Status:** accepted
- **Date:** 2026-08-24

## Context

MCP re-encoding (ADR 0020) covered Claude and OpenCode, both of which have native
MCP config. PI has none — "No MCP. Build an extension" (its README). So PI MCP is not
a config transform; it requires an actual MCP *client*. Research found multiple
production community extensions (`pi-mcp-extension`, etc.) that do this with the
official MCP SDK, and confirmed PI extensions can `import` npm packages.

The choice: depend on a community wrapper, or generate our own. A single-maintainer
third-party package running with full system permissions, on a core capability, cuts
against gyroscope's dependency-light grain — and gyroscope already generates a PI
extension (`gyroscope-personas.ts`, ADR 0015), so an owned MCP extension is the same
pattern. "Our own" still needs the **official** `@modelcontextprotocol/sdk` (the one
dependency any MCP-over-TS path requires); it only removes the wrapper.

## Decision

When `enforce.pi` is set and `mcp` is declared, gyroscope generates, as dedicated
owned files that ride `planMirrors` (written, byte-verified, `--fix`-restored, baked
into the sync script):

- `.pi/mcp.json` — the neutral declaration in the standard `mcpServers` shape (stdio
  first: `command`/`args`/`env`).
- `.pi/extensions/gyroscope-mcp/index.ts` — a minimal MCP-client extension: on load
  it reads `.pi/mcp.json`, connects each stdio server via the official SDK, discovers
  its tools, and `registerTool`-proxies each as `mcp_<server>_<tool>` whose `execute`
  forwards the call. A failing server is skipped, never fatal.
- `.pi/extensions/gyroscope-mcp/package.json` — declares only the official SDK,
  pinned. The teammate runs `npm install` in that dir once (documented, like PI
  `/trust`).

The extension connects **stdio** servers (via `command`) and **remote** servers (via
`url`, streamable-HTTP, with optional `headers`).

**This was validated live** against pi 0.84.3 with BOTH a stdio server (the official
filesystem server) and a remote streamable-HTTP server (the official "everything"
server): the extension loads, connects, registers every tool, and a tool call
round-trips on each transport — and MCP's raw `inputSchema` (JSON Schema) passes
straight to PI's typebox `parameters` with **no shim**, resolving the one risk the
research flagged.

## Consequences

- PI reaches genuine MCP parity from the same neutral `gyroscope.json mcp` block as
  Claude and OpenCode — the "declare once" story is complete across all three
  harnesses in scope.
- gyroscope owns ~60 lines of generated TS plus a one-dependency `package.json`; it
  does not run `npm install` itself (dependency-light) — the skill documents that
  one-time step. The sync script bakes the source files, so a fresh clone reconstructs
  them, then runs `npm install`.
- Low-regret by construction: the generated `.pi/mcp.json` is the standard shape a
  community extension (e.g. `pi-mcp-extension`, pinned) also reads, so swapping to one
  is a docs change, not a rewrite.
- Both transports work (stdio + remote streamable-HTTP). Implementation note for the
  remote path: `StreamableHTTPClientTransport` must be **statically** imported — PI's
  extension loader compiles fine with the static import, and the `--tools <name>`
  check (not "list your tools", which the model formats unreliably) is the sound way
  to verify registration. Running a separate community MCP extension alongside this
  one would double-register — use one.

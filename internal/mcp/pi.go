package mcp

import "github.com/WagnerJust/gyroscope/internal/config"

// PIConfig renders the neutral declaration into `.pi/mcp.json`, the config read by
// the gyroscope-generated MCP-client extension. It is the standard `mcpServers`
// shape (stdio: command/args/env; remote: url/headers). Deterministic for byte-equal
// `check`.
func PIConfig(servers map[string]config.MCPServer) ([]byte, error) {
	m := map[string]any{}
	for name, s := range servers {
		m[name] = piServer(s)
	}
	return encode(map[string]any{"mcpServers": m})
}

func piServer(s config.MCPServer) map[string]any {
	if s.Remote() {
		e := map[string]any{"url": s.URL}
		if len(s.Headers) > 0 {
			e["headers"] = s.Headers
		}
		return e
	}
	e := map[string]any{"command": s.Command}
	if len(s.Args) > 0 {
		e["args"] = s.Args
	}
	if len(s.Env) > 0 {
		e["env"] = s.Env
	}
	return e
}

// PIExtensionPackageJSON is the package.json for the generated MCP-client extension.
// It declares only the OFFICIAL MCP SDK (no community wrapper) — pinned. The teammate
// runs `npm install` in the extension dir once (a documented step, like PI `/trust`).
const PIExtensionPackageJSON = `{
  "name": "gyroscope-mcp",
  "private": true,
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.30.0"
  }
}
`

// PIExtensionIndex is the generated MCP-client extension. PI has no built-in MCP, so
// this bridges it: on load it reads .pi/mcp.json, connects to each stdio server via
// the official SDK, discovers its tools, and registers a proxy tool per MCP tool
// (mcp_<server>_<tool>) whose execute() forwards the call. A failing server is
// skipped, never fatal to startup. This exact shape was validated live against pi
// 0.84.3 + the filesystem MCP server (MCP inputSchema passes straight to PI's
// typebox `parameters`, no shim). Remote (url) transport is a later addition — the
// extension currently connects stdio servers only.
const PIExtensionIndex = `// gyroscope-mcp.ts — managed by gyroscope; do not edit by hand.
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import { readFileSync } from "node:fs";
import { join } from "node:path";

interface Server { command?: string; args?: string[]; env?: Record<string, string>; }

export default async function (pi: ExtensionAPI) {
  let servers: Record<string, Server> = {};
  try {
    const raw = readFileSync(join(process.cwd(), ".pi", "mcp.json"), "utf8");
    servers = (JSON.parse(raw).mcpServers ?? {}) as Record<string, Server>;
  } catch {
    return; // no config: nothing to do
  }
  const clients: Client[] = [];
  for (const [name, s] of Object.entries(servers)) {
    if (!s.command) continue; // stdio only for now
    try {
      const env: Record<string, string> = {};
      for (const [k, v] of Object.entries(process.env)) if (v !== undefined) env[k] = v;
      Object.assign(env, s.env ?? {});
      const client = new Client({ name: "gyroscope-pi", version: "1.0.0" }, { capabilities: {} });
      const transport = new StdioClientTransport({ command: s.command, args: s.args ?? [], env });
      await Promise.race([
        client.connect(transport),
        new Promise((_, reject) => setTimeout(() => reject(new Error("mcp connect timeout")), 20000)),
      ]);
      clients.push(client);
      const listed = await client.listTools();
      for (const t of listed.tools) {
        pi.registerTool({
          name: "mcp_" + name + "_" + t.name,
          label: t.name,
          description: t.description ?? (name + " " + t.name),
          parameters: (t.inputSchema ?? { type: "object", properties: {} }) as any,
          async execute(_id: string, params: unknown) {
            const r = await client.callTool({ name: t.name, arguments: (params ?? {}) as Record<string, unknown> });
            return { content: r.content as any, details: {} };
          },
        });
      }
    } catch {
      // A failing server must not break pi startup; skip it (its own stderr surfaces why).
    }
  }
  pi.on("session_shutdown", async () => {
    for (const c of clients) { try { await c.close(); } catch { /* ignore */ } }
  });
}
`

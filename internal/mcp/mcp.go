// Package mcp re-encodes gyroscope's neutral MCP declaration (config.MCP) into each
// harness's MCP config shape. The schemas genuinely differ — Claude's `mcpServers`
// with stdio/http, OpenCode's `mcp` with local/remote and a merged command array —
// so this is a transform, not a byte-copy (see ADR 0020). Output is deterministic
// (json.Marshal sorts all map keys) so `gyroscope check` can verify it byte-for-byte.
package mcp

import (
	"bytes"
	"encoding/json"

	"github.com/WagnerJust/gyroscope/internal/config"
)

// ClaudeFile renders the declaration into the bytes of Claude's dedicated `.mcp.json`.
func ClaudeFile(servers map[string]config.MCPServer) ([]byte, error) {
	m := map[string]any{}
	for name, s := range servers {
		m[name] = claudeServer(s)
	}
	return encode(map[string]any{"mcpServers": m})
}

func claudeServer(s config.MCPServer) map[string]any {
	if s.Remote() {
		e := map[string]any{"type": "http", "url": s.URL}
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

// OpenCodeServers renders the declaration into the value for opencode.json's `mcp`
// key (one entry per server), for merging into the shared config.
func OpenCodeServers(servers map[string]config.MCPServer) map[string]any {
	m := map[string]any{}
	for name, s := range servers {
		m[name] = openCodeServer(s)
	}
	return m
}

func openCodeServer(s config.MCPServer) map[string]any {
	if s.Remote() {
		e := map[string]any{"type": "remote", "url": s.URL, "enabled": true}
		if len(s.Headers) > 0 {
			e["headers"] = s.Headers
		}
		return e
	}
	cmd := append([]string{s.Command}, s.Args...)
	e := map[string]any{"type": "local", "command": cmd, "enabled": true}
	if len(s.Env) > 0 {
		e["environment"] = s.Env
	}
	return e
}

// encode marshals v as indented JSON with HTML-escaping off (so a `${VAR}` or a `>`
// in a value is written literally), matching how gyroscope writes .claude/settings.json.
func encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

package main

import (
	"fmt"
	"io"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/enforce"
	"github.com/WagnerJust/gyroscope/internal/mcp"
	"github.com/WagnerJust/gyroscope/internal/persona"
)

// planMCPMirrors returns the DEDICATED, gyroscope-owned MCP config files for cfg —
// files gyroscope owns whole (regenerated from the neutral `mcp` declaration),
// gitignorable and baked into the sync script. Currently Claude's `.mcp.json` (PI's
// `.pi/mcp.json` is a later phase). OpenCode's MCP lives in the shared `opencode.json`
// and is merged separately (applyMCP), not owned as a whole file. Empty when no `mcp`
// is declared.
func planMCPMirrors(repoDir string, cfg config.Config) ([]persona.Mirror, error) {
	if len(cfg.MCP) == 0 {
		return nil, nil
	}
	var out []persona.Mirror
	if cfg.Enforce.Claude {
		b, err := mcp.ClaudeFile(cfg.MCP)
		if err != nil {
			return nil, err
		}
		out = append(out, persona.Mirror{Src: "gyroscope.json", Dest: ".mcp.json", Bytes: b})
	}
	return out, nil
}

// applyMCP merges the declared MCP servers into OpenCode's shared opencode.json `mcp`
// key (Claude's dedicated .mcp.json is handled as a mirror). A no-op when no `mcp` is
// declared or the OpenCode adapter is off. Shared by init --apply and check --fix.
func applyMCP(stdout io.Writer, abs string, cfg config.Config) error {
	if len(cfg.MCP) == 0 || !cfg.Enforce.OpenCode {
		return nil
	}
	changed, err := enforce.SetOpenCodeMCP(abs, mcp.OpenCodeServers(cfg.MCP))
	if err != nil {
		return errCannotRun(err)
	}
	if changed {
		fmt.Fprintln(stdout, "merged MCP servers into opencode.json `mcp`")
	}
	return nil
}

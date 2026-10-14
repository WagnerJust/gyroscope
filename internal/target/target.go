// Package target is the doc-target registry: the many tools whose pointer file
// defers to the hub. Enforcement lives in a separate registry (internal/enforce)
// because far fewer tools can run hooks.
package target

import (
	"bytes"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/fsutil"
)

// RoutingLine is the pointer body for every target but Claude Code — one
// identical line so tool-specific files cannot drift.
const RoutingLine = "Before doing anything else, read AGENTS.md and follow its routes.\n"

// claudeHeader opens CLAUDE.md. Claude Code reads AGENTS.md natively only when no
// CLAUDE.md exists, and its docs name a CLAUDE.md that says "read AGENTS.md" in
// words as the thing to replace with an @import — which it loads in full at
// launch, unlike a hook's capped stdout.
const claudeHeader = "<!-- gyroscope: this file imports the hub so Claude Code loads it in full at launch. Edit AGENTS.md, not this file. -->\n"

type Target struct {
	ID   string
	Name string
	Path string // repo-relative pointer path
}

var registry = []Target{
	{"claude", "Claude Code", "CLAUDE.md"},
	{"gemini", "Gemini", "GEMINI.md"},
	{"cursor", "Cursor", ".cursorrules"},
	{"windsurf", "Windsurf", ".windsurfrules"},
	{"cline", "Cline", ".clinerules"},
	{"copilot", "GitHub Copilot", ".github/copilot-instructions.md"},
	{"zed", "Zed", ".rules"},
}

func All() []Target { return append([]Target(nil), registry...) }

func ByID(id string) (Target, bool) {
	for _, t := range registry {
		if t.ID == id {
			return t, true
		}
	}
	return Target{}, false
}

// Body is the pointer content gyroscope writes for t under cfg: the @import body
// for Claude Code (the agents spoke imported only when enabled), the routing
// line for everything else.
func (t Target) Body(cfg config.Config) []byte {
	if t.ID != "claude" {
		return []byte(RoutingLine)
	}
	b := claudeHeader + "@AGENTS.md\n"
	if cfg.Spokes.Agents {
		b += "@docs/agents.md\n"
	}
	return []byte(b)
}

// IsLegacyBody reports whether b is exactly the routing line — the body every
// pointer, CLAUDE.md included, carried before Claude's switched to @imports. Such
// a file is gyroscope-written by construction, so it is safe to upgrade.
func IsLegacyBody(b []byte) bool {
	return bytes.Equal(b, []byte(RoutingLine))
}

func WritePointer(repoDir string, t Target, cfg config.Config, force bool) error {
	return fsutil.WriteGuarded(repoDir, t.Path, t.Body(cfg), force)
}

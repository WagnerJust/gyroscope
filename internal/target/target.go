// Package target is the doc-target registry: the many tools whose pointer file
// gets collapsed to a single routing line back to the hub. Enforcement lives in
// a separate registry (internal/enforce) because far fewer tools can run hooks.
package target

import (
	"github.com/WagnerJust/gyroscope/internal/fsutil"
)

// RoutingLine is the one canonical pointer body — identical across every target
// so tool-specific files cannot drift.
const RoutingLine = "Before doing anything else, read AGENTS.md and follow its routes.\n"

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

func WritePointer(repoDir string, t Target, force bool) error {
	return fsutil.WriteGuarded(repoDir, t.Path, []byte(RoutingLine), force)
}

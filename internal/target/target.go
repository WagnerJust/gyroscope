// Package target is the doc-target registry: the many tools whose pointer file
// gets collapsed to a single routing line back to the hub. Enforcement lives in
// a separate registry (internal/enforce) because far fewer tools can run hooks.
package target

import (
	"fmt"
	"os"
	"path/filepath"
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
	dest := filepath.Join(repoDir, t.Path)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	fh, err := os.OpenFile(dest, flags, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("refusing to overwrite %s (use --force)", t.Path)
		}
		return err
	}
	defer fh.Close()
	_, err = fh.WriteString(RoutingLine)
	return err
}

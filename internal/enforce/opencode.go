package enforce

import (
	"path/filepath"
	"strings"
)

// The OpenCode adapter closes the state-spoke injection gap. OpenCode reads the
// AGENTS.md hub natively, but — unlike Claude's SessionStart hook and PI's
// session_start extension — it does not force-load the non-hub spokes (the resume
// state). OpenCode's `instructions` config key is its always-load mechanism, so
// gyroscope merges the committed non-hub spokes into `opencode.json` `instructions`,
// preserving the user's other entries and keys (the .claude/settings.json merge
// pattern). This is the OpenCode analog of the hook, not a hub injector (the hub is
// native).
//
// Personal `.local/*` spokes are intentionally NOT injected: they are gitignored and
// absent on a fresh clone, and OpenCode's static `instructions` has no missing-file
// tolerance like the hook's `2>/dev/null`. So OpenCode gets the shared resume state
// (TODO.md, the agents spoke, gyroscope.json); the personal `.local/todo.md` is a
// documented gap an OpenCode user can add to their own instructions.

const opencodeConfigPath = "opencode.json"

// OpenCode is the OpenCode enforcement adapter.
type OpenCode struct{}

func (OpenCode) ID() string { return "opencode" }

// opencodeInjectPaths is the non-hub, committed spoke set to force-load: drop
// AGENTS.md (read natively) and any personal .local/ spoke (gitignored / may be
// absent).
func opencodeInjectPaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == "AGENTS.md" || strings.HasPrefix(p, ".local/") {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (OpenCode) PlanLine(paths []string) string {
	return "merge: opencode.json instructions += " + strings.Join(opencodeInjectPaths(paths), ", ")
}

func (OpenCode) Apply(repoDir string, paths []string) (bool, error) {
	return setOpenCodeInstructions(repoDir, opencodeInjectPaths(paths))
}

func (OpenCode) Verify(repoDir string, paths []string) (bool, error) {
	return hasOpenCodeInstructions(repoDir, opencodeInjectPaths(paths))
}

// setOpenCodeInstructions ensures every path in want is present in opencode.json's
// `instructions` array, preserving existing entries and other keys. Returns
// changed=false when all are already present. Reuses the settings JSON read/atomic-
// write (opencode.json is plain JSON; a JSONC file with comments would fail to parse
// and surface as an error, which is honest rather than silently clobbering).
func setOpenCodeInstructions(repoDir string, want []string) (changed bool, err error) {
	path := filepath.Join(repoDir, opencodeConfigPath)
	settings, err := readSettings(path)
	if err != nil {
		return false, err
	}
	cur, _ := settings["instructions"].([]any)
	have := map[string]bool{}
	for _, e := range cur {
		if s, ok := e.(string); ok {
			have[s] = true
		}
	}
	added := false
	for _, w := range want {
		if !have[w] {
			cur = append(cur, w)
			have[w] = true
			added = true
		}
	}
	if !added {
		return false, nil
	}
	settings["instructions"] = cur
	if err := writeSettings(path, settings); err != nil {
		return false, err
	}
	return true, nil
}

// hasOpenCodeInstructions reports whether every path in want is present in
// opencode.json's instructions. A missing file reports false (not installed). The
// read-only inverse of setOpenCodeInstructions.
func hasOpenCodeInstructions(repoDir string, want []string) (bool, error) {
	settings, err := readSettings(filepath.Join(repoDir, opencodeConfigPath))
	if err != nil {
		return false, err
	}
	cur, _ := settings["instructions"].([]any)
	have := map[string]bool{}
	for _, e := range cur {
		if s, ok := e.(string); ok {
			have[s] = true
		}
	}
	for _, w := range want {
		if !have[w] {
			return false, nil
		}
	}
	return true, nil
}

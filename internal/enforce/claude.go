// Package enforce is the enforcement-adapter registry: the few harnesses that can
// run hooks. Each adapter installs a mechanism that makes agents actually read the
// hub. The Claude adapter merges a SessionStart hook into .claude/settings.json,
// preserving any existing settings, and is idempotent.
package enforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SessionStartCommand builds the hook command that cats the given repo-relative
// paths into the session at startup. 2>/dev/null silences the message for a
// missing spoke, and `|| true` forces exit 0: cat still exits non-zero when a
// path is absent (gyroscope.json is missing on the default config; .local/* is
// absent in a fresh clone), which the harness would otherwise report as a failed
// SessionStart hook.
func SessionStartCommand(paths ...string) string {
	return "cat " + strings.Join(paths, " ") + " 2>/dev/null || true"
}

type Claude struct{}

func (Claude) ID() string { return "claude" }

// PlanLine, Apply, and Verify adapt Claude to the Adapter interface: they build
// the hook command from paths and delegate to the existing merge/inspect logic.

func (Claude) PlanLine(paths []string) string {
	return "merge: .claude/settings.json — SessionStart hook: " + SessionStartCommand(paths...)
}

func (c Claude) Apply(repoDir string, paths []string) (bool, error) {
	return c.Install(repoDir, SessionStartCommand(paths...))
}

func (c Claude) Verify(repoDir string, paths []string) (bool, error) {
	return c.HasSessionStart(repoDir, SessionStartCommand(paths...))
}

// Install merges gyroscope's SessionStart hook into repoDir/.claude/settings.json.
// Returns changed=false when the hook is already present.
func (Claude) Install(repoDir string, command string) (changed bool, err error) {
	path := filepath.Join(repoDir, ".claude", "settings.json")
	settings := map[string]any{}
	if b, rerr := os.ReadFile(path); rerr == nil {
		if err := json.Unmarshal(b, &settings); err != nil {
			return false, err
		}
	} else if !os.IsNotExist(rerr) {
		return false, rerr
	}

	// Fail loud rather than silently overwriting a settings file whose "hooks"
	// (or "hooks.SessionStart") isn't the shape we merge into — that would drop
	// the user's existing value on the next write.
	if raw, ok := settings["hooks"]; ok {
		if _, isMap := raw.(map[string]any); !isMap {
			return false, fmt.Errorf("%s: %q is not a JSON object", path, "hooks")
		}
	}
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	if raw, ok := hooks["SessionStart"]; ok {
		if _, isList := raw.([]any); !isList {
			return false, fmt.Errorf("%s: %q is not a JSON array", path, "hooks.SessionStart")
		}
	}
	list, _ := hooks["SessionStart"].([]any)
	if present(list, command) {
		return false, nil
	}
	hooks["SessionStart"] = append(list, map[string]any{
		"hooks": []any{map[string]any{"type": "command", "command": command}},
	})
	settings["hooks"] = hooks

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	// Encode without HTML-escaping so a shell command like "2>/dev/null" is written
	// literally rather than as "2>/dev/null". Encoder.Encode adds a trailing
	// newline for us.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(settings); err != nil {
		return false, err
	}
	// Write to a sibling temp file then rename, so an interrupted write can never
	// truncate the user's existing settings.json — a reader sees old or new, never
	// a partial file.
	tmp := path + ".gyroscope.tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return false, err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return false, err
	}
	return true, nil
}

// HasSessionStart reports whether repoDir/.claude/settings.json already carries a
// SessionStart hook whose command equals command. It is the read-only inverse of
// Install: a missing settings file — or a missing/differently-shaped hooks tree —
// reports false (not present), so a verifier can treat that as nonconformance,
// while a genuine read or parse error is returned so "can't inspect" stays
// distinct from "not there". Reuses present so the match logic lives in one place.
func (Claude) HasSessionStart(repoDir string, command string) (bool, error) {
	path := filepath.Join(repoDir, ".claude", "settings.json")
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var settings map[string]any
	if err := json.Unmarshal(b, &settings); err != nil {
		return false, err
	}
	hooks, _ := settings["hooks"].(map[string]any)
	list, _ := hooks["SessionStart"].([]any)
	return present(list, command), nil
}

func present(list []any, command string) bool {
	for _, e := range list {
		m, _ := e.(map[string]any)
		inner, _ := m["hooks"].([]any)
		for _, h := range inner {
			hm, _ := h.(map[string]any)
			if cmd, _ := hm["command"].(string); cmd == command {
				return true
			}
		}
	}
	return false
}

// Package enforce is the enforcement-adapter registry: the few harnesses that can
// run hooks. Each adapter installs a mechanism that delivers the repo's rules and
// working state into a session. The Claude adapter merges a SessionStart hook that
// injects the state files within a budget into .claude/settings.json, preserving
// any existing settings, and is idempotent.
package enforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Claude struct{}

func (Claude) ID() string { return "claude" }

// PlanLine, Apply, and Verify adapt Claude to the Adapter interface. The hook
// carries p.State only: the rules reach Claude Code through CLAUDE.md's @imports,
// which it loads in full at launch, so catting them here would only spend the
// hook's 10,000-character budget twice. With no state there is no hook at all.

func (Claude) PlanLine(p Paths) string {
	if len(p.State) == 0 {
		return "merge: .claude/settings.json — no SessionStart hook (no state spokes enabled); any gyroscope hook is removed"
	}
	return fmt.Sprintf("merge: .claude/settings.json — SessionStart hook injects state within %d bytes: %s",
		StateBudget, strings.Join(p.State, " "))
}

func (c Claude) Apply(repoDir string, p Paths) (bool, error) {
	if len(p.State) == 0 {
		return c.remove(repoDir)
	}
	return c.Install(repoDir, SessionStartCommand(p.State))
}

func (c Claude) Verify(repoDir string, p Paths) (bool, error) {
	if len(p.State) == 0 {
		list, err := readSessionStart(repoDir)
		if err != nil {
			return false, err
		}
		_, pruned := pruneGyroscope(list, "")
		return !pruned, nil
	}
	return c.HasSessionStart(repoDir, SessionStartCommand(p.State))
}

// isGyroscopeHook reports whether a SessionStart command is one gyroscope wrote:
// the legacy `cat AGENTS.md ...` shape, or any version carrying hookTag.
func isGyroscopeHook(command string) bool {
	return strings.HasPrefix(command, "cat AGENTS.md") || strings.Contains(command, hookTag)
}

// pruneGyroscope returns list without gyroscope's own inner hooks, except one
// whose command equals keep. An entry group left with no hooks is dropped; every
// other hook and group key (matcher, ...) is untouched. pruned reports whether
// anything was removed.
func pruneGyroscope(list []any, keep string) (out []any, pruned bool) {
	out = make([]any, 0, len(list))
	for _, e := range list {
		m, ok := e.(map[string]any)
		inner, isList := m["hooks"].([]any)
		if !ok || !isList {
			out = append(out, e)
			continue
		}
		kept := make([]any, 0, len(inner))
		for _, h := range inner {
			hm, _ := h.(map[string]any)
			if cmd, _ := hm["command"].(string); cmd != keep && isGyroscopeHook(cmd) {
				pruned = true
				continue
			}
			kept = append(kept, h)
		}
		if len(kept) == 0 && len(inner) > 0 {
			continue
		}
		if len(kept) != len(inner) {
			// Copy rather than mutate the decoded group in place.
			g := make(map[string]any, len(m))
			for k, v := range m {
				g[k] = v
			}
			g["hooks"] = kept
			e = g
		}
		out = append(out, e)
	}
	return out, pruned
}

// sessionStartTree loads settings and returns its hooks object and SessionStart
// list, failing loud when either is not the shape gyroscope merges into —
// silently replacing it would drop the user's existing value on the next write.
func sessionStartTree(path string) (settings, hooks map[string]any, list []any, err error) {
	settings, err = readSettings(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if raw, ok := settings["hooks"]; ok {
		if _, isMap := raw.(map[string]any); !isMap {
			return nil, nil, nil, fmt.Errorf("%s: %q is not a JSON object", path, "hooks")
		}
	}
	hooks, _ = settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	if raw, ok := hooks["SessionStart"]; ok {
		if _, isList := raw.([]any); !isList {
			return nil, nil, nil, fmt.Errorf("%s: %q is not a JSON array", path, "hooks.SessionStart")
		}
	}
	list, _ = hooks["SessionStart"].([]any)
	return settings, hooks, list, nil
}

// Install merges gyroscope's SessionStart hook into repoDir/.claude/settings.json,
// replacing any earlier gyroscope hook (legacy or tagged) rather than appending
// beside it, and leaving every other hook untouched. Returns changed=false when
// the hook is already present and nothing stale needed removing.
func (Claude) Install(repoDir string, command string) (changed bool, err error) {
	path := filepath.Join(repoDir, ".claude", "settings.json")
	settings, hooks, list, err := sessionStartTree(path)
	if err != nil {
		return false, err
	}
	list, pruned := pruneGyroscope(list, command)
	if !present(list, command) {
		list = append(list, map[string]any{
			"hooks": []any{map[string]any{"type": "command", "command": command}},
		})
	} else if !pruned {
		return false, nil
	}
	hooks["SessionStart"] = list
	settings["hooks"] = hooks

	if err := writeSettings(path, settings); err != nil {
		return false, err
	}
	return true, nil
}

// remove drops every gyroscope SessionStart hook from repoDir/.claude/settings.json
// (the no-state case). A missing file, or one with nothing to remove, is left
// untouched and reports changed=false.
func (Claude) remove(repoDir string) (changed bool, err error) {
	path := filepath.Join(repoDir, ".claude", "settings.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}
	settings, hooks, list, err := sessionStartTree(path)
	if err != nil {
		return false, err
	}
	list, pruned := pruneGyroscope(list, "")
	if !pruned {
		return false, nil
	}
	hooks["SessionStart"] = list
	settings["hooks"] = hooks
	if err := writeSettings(path, settings); err != nil {
		return false, err
	}
	return true, nil
}

// readSettings loads a .claude/settings.json object, returning an empty map when the
// file is absent. A malformed file is an error rather than a silent reset — gyroscope
// merges into the user's settings and must never clobber an unparseable one.
func readSettings(path string) (map[string]any, error) {
	settings := map[string]any{}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return settings, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// writeSettings encodes settings back to path, creating .claude/ as needed. It
// disables HTML-escaping so a shell fragment like "2>/dev/null" is written literally
// (not "2>/dev/null"), and writes via a sibling temp file + rename so an interrupted
// write can never truncate the user's settings — a reader sees the old file or the new.
func writeSettings(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(settings); err != nil {
		return err
	}
	tmp := path + ".gyroscope.tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// SetCoAuthoredBy sets the top-level includeCoAuthoredBy key in
// repoDir/.claude/settings.json, preserving every other setting. This is Claude
// Code's lever for the "Co-Authored-By: Claude" git trailer and the "Generated with
// Claude Code" line on commits and PRs. gyroscope writes the key only to SUPPRESS
// attribution (include=false); it never forces it true, leaving Claude's own default
// otherwise. Returns changed=false when the key already holds include.
func SetCoAuthoredBy(repoDir string, include bool) (changed bool, err error) {
	path := filepath.Join(repoDir, ".claude", "settings.json")
	settings, err := readSettings(path)
	if err != nil {
		return false, err
	}
	if cur, ok := settings["includeCoAuthoredBy"].(bool); ok && cur == include {
		return false, nil
	}
	settings["includeCoAuthoredBy"] = include
	if err := writeSettings(path, settings); err != nil {
		return false, err
	}
	return true, nil
}

// CoAuthoredBy reports the includeCoAuthoredBy setting in
// repoDir/.claude/settings.json: value is its boolean, present whether the key exists
// as a bool. A missing settings file reports present=false. Read-only inverse of
// SetCoAuthoredBy.
func CoAuthoredBy(repoDir string) (value, present bool, err error) {
	settings, err := readSettings(filepath.Join(repoDir, ".claude", "settings.json"))
	if err != nil {
		return false, false, err
	}
	v, ok := settings["includeCoAuthoredBy"].(bool)
	return v, ok, nil
}

// HasSessionStart reports whether repoDir/.claude/settings.json already carries a
// SessionStart hook whose command equals command. It is the read-only inverse of
// Install: a missing settings file — or a missing/differently-shaped hooks tree —
// reports false (not present), so a verifier can treat that as nonconformance,
// while a genuine read or parse error is returned so "can't inspect" stays
// distinct from "not there". Reuses present so the match logic lives in one place.
func (Claude) HasSessionStart(repoDir string, command string) (bool, error) {
	list, err := readSessionStart(repoDir)
	if err != nil {
		return false, err
	}
	return present(list, command), nil
}

// readSessionStart returns the SessionStart list from repoDir/.claude/settings.json
// for read-only inspection: a missing file or a missing/differently-shaped hooks
// tree yields an empty list, a read or parse error is returned.
func readSessionStart(repoDir string) ([]any, error) {
	b, err := os.ReadFile(filepath.Join(repoDir, ".claude", "settings.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var settings map[string]any
	if err := json.Unmarshal(b, &settings); err != nil {
		return nil, err
	}
	hooks, _ := settings["hooks"].(map[string]any)
	list, _ := hooks["SessionStart"].([]any)
	return list, nil
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

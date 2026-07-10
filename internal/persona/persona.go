// Package persona mirrors a repo's canonical personas from docs/agents/ into
// .claude/agents/ so Claude Code registers them as dispatchable subagents.
//
// The personas are authored (by the /gyroscope skill) in valid Claude subagent
// format — YAML frontmatter (name/description/tools/model) plus a system-prompt
// body — but they live in docs/agents/, which Claude Code does NOT scan for
// subagents. Claude registers subagents only from .claude/agents/ (project) or
// ~/.claude/agents/ (user). So docs/agents/ stays the canonical, hub-routed
// source and each valid persona is COPIED, byte-for-byte, into .claude/agents/
// (see ADR 0010). Copy, not symlink: robust across clones and OSes, and any drift
// is caught by `gyroscope check`.
//
// The binary mirrors persona BYTES; it never authors persona content — consistent
// with "the binary copies bytes, it does not render." Whether the mirror runs is
// gated by the caller (personas == on AND enforce.claude); this package only knows
// how to copy.
package persona

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Mirror is a single planned copy: the repo-relative source persona under
// docs/agents/ and the repo-relative .claude/agents/<name>.md destination it is
// registered as. Bytes are the exact source content written verbatim to Dest.
type Mirror struct {
	Src   string
	Dest  string
	Bytes []byte
}

// Plan lists the mirror copies gyroscope would write for repoDir: one per VALID
// persona under docs/agents/ — a *.md file (other than README.md) whose YAML
// frontmatter carries a `name:`. The destination filename is that frontmatter
// name, so `.claude/agents/<name>.md` matches the registered subagent type even
// when the source file is named differently. A missing docs/agents/ yields an
// empty plan (no personas to register), not an error.
func Plan(repoDir string) ([]Mirror, error) {
	agentsDir := filepath.Join(repoDir, "docs", "agents")
	entries, err := os.ReadDir(agentsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var mirrors []Mirror
	for _, e := range entries {
		if e.IsDir() || e.Name() == "README.md" || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(agentsDir, e.Name()))
		if err != nil {
			return nil, err
		}
		name, ok := frontmatterName(b)
		if !ok {
			// Not a valid persona (no `name:` frontmatter): skip it. Only files
			// Claude can register as a subagent get mirrored.
			continue
		}
		mirrors = append(mirrors, Mirror{
			Src:   filepath.Join("docs", "agents", e.Name()),
			Dest:  filepath.Join(".claude", "agents", name+".md"),
			Bytes: b,
		})
	}
	return mirrors, nil
}

// frontmatterName extracts the `name:` value from a persona's YAML frontmatter —
// the block delimited by a leading `---` line and the next `---` line. It returns
// ("", false) when there is no frontmatter or no `name:` key, which marks the file
// as not a valid persona. Only the small subset of YAML personas actually use is
// parsed (a `key: value` line); no YAML dependency is pulled in.
func frontmatterName(b []byte) (string, bool) {
	s := string(b)
	// Frontmatter must be the first thing in the file.
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return "", false
	}
	lines := strings.Split(s, "\n")
	// lines[0] is the opening "---"; scan until the closing "---".
	for _, ln := range lines[1:] {
		t := strings.TrimRight(ln, "\r")
		if t == "---" {
			break // end of frontmatter, no name found
		}
		key, val, ok := strings.Cut(t, ":")
		if !ok {
			continue
		}
		if strings.TrimSpace(key) == "name" {
			name := strings.TrimSpace(val)
			if name == "" {
				return "", false
			}
			return name, true
		}
	}
	return "", false
}

// Copy writes each planned mirror to repoDir, creating .claude/agents/ as needed,
// and returns the destinations written. It writes NOTHING (and creates no
// directory) when there are no valid personas, so a persona-less repo stays clean.
//
// gyroscope OWNS the persona-named mirror files: this is a generated mirror, like
// the SessionStart hook is idempotently re-applied, so it must overwrite on drift.
// That is the one deliberate non-WriteGuarded write in the codebase — WriteGuarded's
// O_EXCL would refuse the re-mirror. The write is atomic (temp file + rename) so a
// reader never sees a partial file and an interrupted write can't truncate a mirror.
// It never touches .claude/settings.json or any non-persona .claude/ content.
func Copy(repoDir string) (written []string, err error) {
	mirrors, err := Plan(repoDir)
	if err != nil {
		return nil, err
	}
	for _, m := range mirrors {
		if err := writeMirror(filepath.Join(repoDir, m.Dest), m.Bytes); err != nil {
			return written, err
		}
		written = append(written, m.Dest)
	}
	return written, nil
}

// writeMirror writes content to dest atomically (temp + rename), creating the
// parent directory. Overwrites unconditionally — the mirror is gyroscope-owned.
func writeMirror(dest string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".gyroscope.tmp"
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("mirror %s: %w", dest, err)
	}
	return nil
}

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

// Plan lists the Claude mirror copies gyroscope would write for repoDir — the
// back-compatible default target. See PlanFor for the per-harness generalization.
func Plan(repoDir string) ([]Mirror, error) {
	return PlanFor(repoDir, Claude)
}

// PlanFor lists the mirror files gyroscope would write for repoDir to register its
// personas as subagents in target t's registry: one per VALID persona under
// docs/agents/ — a *.md file (other than README.md) whose YAML frontmatter carries
// a `name:`. The destination path and encoding come from t, so the SAME canonical
// source registers into each harness's own format. The destination filename is the
// frontmatter name, so it matches the registered subagent id even when the source
// file is named differently. A missing docs/agents/ yields an empty plan (no
// personas to register), not an error.
func PlanFor(repoDir string, t Target) ([]Mirror, error) {
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
		p, ok := parse(b)
		if !ok {
			// Not a valid persona (no `name:` frontmatter): skip it. Only files a
			// harness can register as a subagent get mirrored.
			continue
		}
		mirrors = append(mirrors, Mirror{
			Src:   filepath.Join("docs", "agents", e.Name()),
			Dest:  t.dest(p.Name),
			Bytes: t.render(p),
		})
	}
	return mirrors, nil
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
	return Write(repoDir, mirrors)
}

// Write copies each already-planned mirror under repoDir and returns the
// destinations written. Callers that have already Plan-ned (init, check) reuse it
// so the plan is not re-derived; Copy is the plan-then-write convenience. It
// writes nothing (and creates no directory) for an empty plan.
func Write(repoDir string, mirrors []Mirror) (written []string, err error) {
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

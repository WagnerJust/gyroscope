package enforce

import (
	"os"
	"path/filepath"

	"github.com/WagnerJust/gyroscope/internal/fsutil"
)

// The Cursor adapter is the force-inject analog for Cursor. Cursor reads AGENTS.md
// natively, but that is passive; a project rule under .cursor/rules/ with
// `alwaysApply: true` is injected into EVERY model request, which is the real
// enforcement primitive (parallel to Claude's SessionStart hook / PI's session_start
// extension). It supersedes the passive, deprecated `.cursorrules` pointer for
// agent-mode Cursor; gyroscope keeps that pointer too for backward compatibility.
//
// Unlike the Claude/PI adapters, the rule is STATIC — a Cursor rule is injected text
// and cannot cat live files, so it force-injects the directive to read the hub
// rather than the hub's contents. The hub's own routes carry the agent to the
// state/resume spokes from there.

// cursorRulePath is the auto-applied project rule Cursor loads. `.mdc` is Cursor's
// rule format (YAML frontmatter + markdown body).
const cursorRulePath = ".cursor/rules/gyroscope-hub.mdc"

// cursorRuleBody is the managed always-apply rule. The body matches target.RoutingLine
// so the force-inject directive reads identically to the pointer files.
const cursorRuleBody = "---\n" +
	"description: Always read the AGENTS.md hub before acting\n" +
	"alwaysApply: true\n" +
	"---\n" +
	"Before doing anything else, read AGENTS.md and follow its routes.\n"

// Cursor is the Cursor enforcement adapter.
type Cursor struct{}

func (Cursor) ID() string { return "cursor" }

// PlanLine, Apply, Verify ignore paths: the rule is static (a Cursor rule cannot
// read files live), so it does not vary with the injected spoke set.

func (Cursor) PlanLine(paths []string) string {
	return "write: " + cursorRulePath + " (Cursor always-apply hub rule)"
}

func (Cursor) Apply(repoDir string, paths []string) (bool, error) {
	dest := filepath.Join(repoDir, cursorRulePath)
	if b, err := os.ReadFile(dest); err == nil {
		if string(b) == cursorRuleBody {
			return false, nil
		}
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := fsutil.WriteGuarded(repoDir, cursorRulePath, []byte(cursorRuleBody), true); err != nil {
		return false, err
	}
	return true, nil
}

func (Cursor) Verify(repoDir string, paths []string) (bool, error) {
	b, err := os.ReadFile(filepath.Join(repoDir, cursorRulePath))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return string(b) == cursorRuleBody, nil
}

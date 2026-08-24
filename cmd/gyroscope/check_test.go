package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/WagnerJust/gyroscope/internal/enforce"
	"github.com/WagnerJust/gyroscope/internal/standard"
)

// placeholderRE matches a fill-once {{...}} scaffold marker, including one that
// spans multiple lines. Tests use it to stand in for the /gyroscope skill's
// interview-and-fill pass, so `check` sees the conformant, post-fill state a real
// committed repo has — a raw `init` leaves these markers unfilled by design.
var placeholderRE = regexp.MustCompile(`(?s)\{\{.*?\}\}`)

// fillPlaceholders strips every {{...}} marker from files under repo that carry
// one, the read-only stand-in for the skill filling gyroscope's scaffolds.
func fillPlaceholders(t *testing.T, repo string) {
	t.Helper()
	err := filepath.WalkDir(repo, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Contains(b, []byte("{{")) {
			return nil
		}
		return os.WriteFile(path, placeholderRE.ReplaceAll(b, []byte("filled")), 0o644)
	})
	if err != nil {
		t.Fatalf("fillPlaceholders: %v", err)
	}
}

// exitCodeOf reports the exit code a run() error maps to (0 for nil).
func exitCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var ee *exitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *exitError, got %T: %v", err, err)
	}
	return ee.code
}

// initAndFill applies the standard into a fresh temp repo and fills its
// scaffolds, yielding the conformant baseline a committed repo has.
func initAndFill(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)
	return dir
}

func TestCheckPassesOnFilledRepo(t *testing.T) {
	dir := initAndFill(t)
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass on a filled repo, got %v\nstdout: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "conformant") {
		t.Fatalf("expected a conformant confirmation, got: %s", out.String())
	}
}

func TestCheckFailsOnMissingSpoke(t *testing.T) {
	dir := initAndFill(t)
	if err := os.Remove(filepath.Join(dir, "CONTEXT.md")); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("expected exit %d (drift), got %d (err %v)", exitDrift, code, err)
	}
	if !strings.Contains(out.String(), "CONTEXT.md") {
		t.Fatalf("drift report should name the missing spoke, got: %s", out.String())
	}
}

func TestCheckFailsOnCorruptPointer(t *testing.T) {
	dir := initAndFill(t)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("junk — no routing line here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("expected exit %d (drift), got %d (err %v)", exitDrift, code, err)
	}
	if !strings.Contains(out.String(), "CLAUDE.md") {
		t.Fatalf("drift report should name the corrupt pointer, got: %s", out.String())
	}
}

func TestCheckFailsOnUnfilledPlaceholder(t *testing.T) {
	dir := initAndFill(t)
	// Re-introduce a fill-once marker into a placeholder-bearing spoke.
	p := filepath.Join(dir, "docs", "agents.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, append(b, []byte("\n{{unfilled}}\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err = run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("expected exit %d (drift), got %d (err %v)", exitDrift, code, err)
	}
	if !strings.Contains(out.String(), "docs/agents.md") {
		t.Fatalf("drift report should name the file with the leftover placeholder, got: %s", out.String())
	}
}

// setTODO overwrites the repo's TODO.md, keeping the rest of the filled repo
// conformant so a test isolates the archive-nudge heuristic.
func setTODO(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "TODO.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The archive nudge is a SOFT note, not drift: an unarchived done item is untidy
// but not a structural nonconformance, so a repo over the threshold still exits 0
// (conformant) while printing the nudge. Over-threshold branch.
func TestCheckArchiveNudgeOverThresholdIsSoftNote(t *testing.T) {
	dir := initAndFill(t)
	// 6 completed items in TODO.md — over the threshold of 5.
	setTODO(t, dir, "# TODO\n\n## Now\n- [ ] open one\n\n## Done\n"+
		"- [x] a\n- [x] b\n- [x] c\n- [x] d\n- [x] e\n- [x] f\n")
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("archive nudge must stay a soft note (exit 0), got %v\n%s", err, out.String())
	}
	s := out.String()
	if !strings.Contains(s, "conformant") {
		t.Fatalf("over-threshold repo should still be conformant, got: %s", s)
	}
	if !strings.Contains(s, "DONE.md") || !strings.Contains(strings.ToLower(s), "archive") {
		t.Fatalf("expected an archive-to-DONE.md nudge, got: %s", s)
	}
}

// Under-threshold branch: a small number of completed items in TODO.md draws no
// nudge (and stays conformant).
func TestCheckArchiveNudgeUnderThresholdIsSilent(t *testing.T) {
	dir := initAndFill(t)
	// 2 completed items — under the threshold.
	setTODO(t, dir, "# TODO\n\n## Now\n- [ ] open one\n- [x] a\n- [x] b\n")
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass, got %v\n%s", err, out.String())
	}
	s := out.String()
	if !strings.Contains(s, "conformant") {
		t.Fatalf("under-threshold repo should be conformant, got: %s", s)
	}
	if strings.Contains(s, "archive") || strings.Contains(s, "DONE.md") {
		t.Fatalf("under-threshold repo must draw no archive nudge, got: %s", s)
	}
}

// When the Claude adapter is on but .claude/ is gitignored, the SessionStart hook
// (and, with personas on, the .claude/agents/ mirror) are local-only — a teammate
// who clones gets the tracked docs but no auto-injection. check emits a soft note
// advising a targeted negation, without failing (gitignoring .claude/ is legit).
func TestCheckNotesGitignoredClaudeDir(t *testing.T) {
	dir := initAndFill(t)
	gi := filepath.Join(dir, ".gitignore")
	base, err := os.ReadFile(gi)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gi, append(append([]byte{}, base...), []byte("\n.claude/\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("gitignored .claude/ is advisory, must stay exit 0, got %v\n%s", err, out.String())
	}
	s := out.String()
	if !strings.Contains(s, "conformant") {
		t.Fatalf("should still be conformant, got: %s", s)
	}
	if !strings.Contains(s, "won't ship") || !strings.Contains(s, "!.claude/settings.json") {
		t.Fatalf("expected a ship-to-team note advising the negation, got: %s", s)
	}

	// A negation re-including the hook cancels the note.
	if err := os.WriteFile(gi, append(append([]byte{}, base...), []byte("\n.claude/\n!.claude/settings.json\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "won't ship") {
		t.Fatalf("a `!.claude/settings.json` negation should cancel the note, got: %s", out.String())
	}
}

// initAndFillPI is initAndFill with PI enforcement enabled.
func initAndFillPI(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"pi":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)
	return dir
}

func TestCheckPassesWithPIEnabled(t *testing.T) {
	dir := initAndFillPI(t)
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass with PI enabled, got %v\n%s", err, out.String())
	}
}

func TestCheckDriftsWhenPIExtensionMissing(t *testing.T) {
	dir := initAndFillPI(t)
	if err := os.Remove(filepath.Join(dir, ".pi", "extensions", "gyroscope-context.ts")); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("removing the PI extension should drift, got code %d (%v)", code, err)
	}
	if !strings.Contains(out.String(), "pi") {
		t.Fatalf("drift should name the pi adapter, got: %s", out.String())
	}
}

// A valid persona (frontmatter name + description + body) for the mirror tests.
const mirrorPersona = "---\n" +
	"name: backend-architect\n" +
	`description: Use for "design" work, or review.` + "\n" +
	"tools: Read, Grep\n" +
	"---\n\n# Backend Architect\n\nYou design things.\n"

// With OpenCode and Codex enabled alongside Claude, init registers each persona in
// ALL THREE native subagent registries (byte-copy for Claude, re-encoded for the
// others), check verifies them, and a removed mirror is drift.
func TestCheckRegistersPersonasForOpenCodeAndCodex(t *testing.T) {
	dir := t.TempDir()
	cfg := `{"spokes":{"personas":"on"},"enforce":{"claude":true,"opencode":true,"codex":true}}`
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	// The persona must exist before init so init mirrors it into every registry.
	if err := os.MkdirAll(filepath.Join(dir, "docs", "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "agents", "backend-architect.md"), []byte(mirrorPersona), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	claude := filepath.Join(dir, ".claude", "agents", "backend-architect.md")
	opencode := filepath.Join(dir, ".opencode", "agents", "backend-architect.md")
	codex := filepath.Join(dir, ".codex", "agents", "backend-architect.toml")
	for _, p := range []string{claude, opencode, codex} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected registered mirror %s: %v", p, err)
		}
	}
	// Spot-check the re-encoded formats.
	if b, _ := os.ReadFile(opencode); !strings.Contains(string(b), "mode: subagent") {
		t.Errorf("opencode mirror missing `mode: subagent`:\n%s", b)
	}
	if b, _ := os.ReadFile(codex); !strings.Contains(string(b), "developer_instructions") {
		t.Errorf("codex mirror missing developer_instructions:\n%s", b)
	}

	// Conformant with all three present.
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass with all mirrors present, got %v\n%s", err, out.String())
	}

	// Removing the OpenCode mirror is drift naming .opencode.
	if err := os.Remove(opencode); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	err := run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("missing opencode mirror should drift, got code %d (%v)", code, err)
	}
	if !strings.Contains(out.String(), ".opencode") {
		t.Fatalf("drift should name the opencode mirror, got: %s", out.String())
	}

	// check --fix re-registers it → conformant again.
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", "--fix", dir}, &out, &errb); err != nil {
		t.Fatalf("check --fix: %v\n%s", err, out.String())
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass after --fix re-mirrored, got %v\n%s", err, out.String())
	}
}

// PI has no native subagent registry, so persona parity is a generated extension
// that registers a subagent tool per persona. With PI enabled and personas on, init
// writes .pi/extensions/gyroscope-personas.ts, check verifies it, and removing it is
// drift that --fix restores.
func TestCheckRegistersPersonasForPI(t *testing.T) {
	dir := t.TempDir()
	cfg := `{"spokes":{"personas":"on"},"enforce":{"claude":true,"pi":true}}`
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "docs", "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "agents", "backend-architect.md"), []byte(mirrorPersona), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	ext := filepath.Join(dir, ".pi", "extensions", "gyroscope-personas.ts")
	b, err := os.ReadFile(ext)
	if err != nil {
		t.Fatalf("expected PI persona extension: %v", err)
	}
	if !strings.Contains(string(b), "pi.registerTool(") || !strings.Contains(string(b), `"name": "backend-architect"`) {
		t.Fatalf("PI extension does not register the persona as a tool:\n%s", b)
	}

	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass with the PI extension present, got %v\n%s", err, out.String())
	}

	// Remove it → drift.
	if err := os.Remove(ext); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("missing PI persona extension should drift, got code %d", code)
	}
	if !strings.Contains(out.String(), "gyroscope-personas.ts") {
		t.Fatalf("drift should name the PI persona extension, got: %s", out.String())
	}

	// --fix restores it.
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", "--fix", dir}, &out, &errb); err != nil {
		t.Fatalf("check --fix: %v\n%s", err, out.String())
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass after --fix restored the PI extension, got %v\n%s", err, out.String())
	}
}

// The Cursor adapter installs a .cursor/rules/ always-apply rule (the force-inject
// analog for Cursor); init writes it, check verifies it, removing it drifts, --fix
// restores it.
func TestCheckCursorAdapter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"cursor":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	rule := filepath.Join(dir, ".cursor", "rules", "gyroscope-hub.mdc")
	if b, err := os.ReadFile(rule); err != nil || !strings.Contains(string(b), "alwaysApply: true") {
		t.Fatalf("expected an always-apply Cursor rule, err=%v", err)
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass with the Cursor rule present, got %v\n%s", err, out.String())
	}

	if err := os.Remove(rule); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("removing the Cursor rule should drift, got code %d", code)
	}
	if !strings.Contains(out.String(), "cursor") {
		t.Fatalf("drift should name the cursor adapter, got: %s", out.String())
	}

	out.Reset()
	errb.Reset()
	if err := run([]string{"check", "--fix", dir}, &out, &errb); err != nil {
		t.Fatalf("check --fix should restore the rule, got %v\n%s", err, out.String())
	}
	if b, err := os.ReadFile(rule); err != nil || !strings.Contains(string(b), "alwaysApply: true") {
		t.Fatalf("--fix did not restore the Cursor rule (err=%v)", err)
	}
}

// With sync enabled, init writes a committed reconstruction script; check verifies
// it; and — the real proof — running it on a fresh clone (mirrors deleted)
// reconstructs them byte-for-byte, idempotently.
func TestSyncScriptReconstructsMirrors(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	dir := t.TempDir()
	cfg := `{"sync":true,"spokes":{"personas":"on"},"enforce":{"claude":true,"opencode":true}}`
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "docs", "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "agents", "backend-architect.md"), []byte(mirrorPersona), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	script := filepath.Join(dir, "gyroscope-sync.sh")
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("sync script not written: %v", err)
	}
	// check passes with the script current.
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass with the sync script current, got %v\n%s", err, out.String())
	}

	// Simulate a fresh clone: the generated mirrors were gitignored and are absent.
	claude := filepath.Join(dir, ".claude", "agents", "backend-architect.md")
	opencode := filepath.Join(dir, ".opencode", "agents", "backend-architect.md")
	want, err := os.ReadFile(claude)
	if err != nil {
		t.Fatal(err)
	}
	wantOC, err := os.ReadFile(opencode)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, ".claude", "agents")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, ".opencode")); err != nil {
		t.Fatal(err)
	}

	// Run the committed script — it must reconstruct both mirrors byte-for-byte.
	cmd := exec.Command("sh", "gyroscope-sync.sh")
	cmd.Dir = dir
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gyroscope-sync.sh failed: %v\n%s", err, b)
	}
	got, err := os.ReadFile(claude)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("claude mirror not reconstructed byte-for-byte (err=%v)", err)
	}
	gotOC, err := os.ReadFile(opencode)
	if err != nil || !bytes.Equal(gotOC, wantOC) {
		t.Fatalf("opencode mirror not reconstructed byte-for-byte (err=%v)", err)
	}

	// Idempotent: a second run does not error and does not alter the reconstructed files.
	cmd2 := exec.Command("sh", "gyroscope-sync.sh")
	cmd2.Dir = dir
	if b, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("second sync run failed: %v\n%s", err, b)
	}
	if got2, _ := os.ReadFile(claude); !bytes.Equal(got2, want) {
		t.Fatal("second run altered an already-present mirror (not skip-if-exists)")
	}
}

const testCommand = "---\n" +
	"description: Fix a GitHub issue\n" +
	"argument-hint: [issue-number]\n" +
	"allowed-tools: Bash(git *)\n" +
	"---\nFix issue $ARGUMENTS.\n"

// A canonical docs/commands/ command re-encodes into each enabled harness's command
// location — independent of the personas state — and check verifies each byte-equal.
func TestCheckGeneratesCommandsPerHarness(t *testing.T) {
	dir := t.TempDir()
	// personas OFF on purpose: commands must not depend on the persona state.
	cfg := `{"spokes":{"personas":"off"},"enforce":{"claude":true,"opencode":true,"pi":true}}`
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "docs", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "commands", "fix.md"), []byte(testCommand), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	claude := filepath.Join(dir, ".claude", "commands", "fix.md")
	opencode := filepath.Join(dir, ".opencode", "commands", "fix.md")
	pi := filepath.Join(dir, ".pi", "prompts", "fix.md")
	for _, p := range []string{claude, opencode, pi} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected generated command %s: %v", p, err)
		}
	}
	// OpenCode drops argument-hint; PI keeps it; both keep the body.
	if b, _ := os.ReadFile(opencode); strings.Contains(string(b), "argument-hint") {
		t.Errorf("opencode command should drop argument-hint:\n%s", b)
	}
	if b, _ := os.ReadFile(pi); !strings.Contains(string(b), "argument-hint: [issue-number]") {
		t.Errorf("pi command should keep argument-hint:\n%s", b)
	}
	if b, _ := os.ReadFile(pi); !strings.Contains(string(b), "Fix issue $ARGUMENTS.") {
		t.Errorf("pi command must keep the body verbatim:\n%s", b)
	}

	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass with commands generated, got %v\n%s", err, out.String())
	}

	// Remove one → drift; --fix restores.
	if err := os.Remove(opencode); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("missing command should drift, got code %d", code)
	}
	if !strings.Contains(out.String(), ".opencode/commands") {
		t.Fatalf("drift should name the missing command, got: %s", out.String())
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", "--fix", dir}, &out, &errb); err != nil {
		t.Fatalf("check --fix: %v\n%s", err, out.String())
	}
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass after --fix restored the command, got %v", err)
	}
}

// enforce.opencode installs the OpenCode adapter: it merges the non-hub state
// spokes into opencode.json `instructions`. init writes it, check verifies it,
// removing the key drifts, --fix restores it.
func TestCheckOpenCodeInstructionsAdapter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"opencode":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	oc := filepath.Join(dir, "opencode.json")
	b, err := os.ReadFile(oc)
	if err != nil || !strings.Contains(string(b), "instructions") || !strings.Contains(string(b), "TODO.md") {
		t.Fatalf("opencode.json should carry instructions incl TODO.md, got err=%v\n%s", err, b)
	}
	if strings.Contains(string(b), "AGENTS.md") {
		t.Errorf("AGENTS.md must not be in instructions (native read):\n%s", b)
	}

	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass with the OpenCode adapter installed, got %v\n%s", err, out.String())
	}

	// Drop the instructions key → drift; --fix restores.
	if err := os.WriteFile(oc, []byte(`{"model":"anthropic/x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("removing opencode instructions should drift, got code %d", code)
	}
	if !strings.Contains(out.String(), "opencode") {
		t.Fatalf("drift should name the opencode adapter, got: %s", out.String())
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", "--fix", dir}, &out, &errb); err != nil {
		t.Fatalf("check --fix: %v\n%s", err, out.String())
	}
	// --fix preserved the user's model key while re-adding instructions.
	if b, _ := os.ReadFile(oc); !strings.Contains(string(b), "anthropic/x") || !strings.Contains(string(b), "TODO.md") {
		t.Fatalf("--fix should merge, preserving the user key:\n%s", b)
	}
}

// A neutral mcp block re-encodes into Claude's dedicated .mcp.json (a gyroscope-owned
// mirror) and merges into OpenCode's shared opencode.json `mcp`. check verifies both;
// drift in either is caught and --fix restores it.
func TestCheckMCPClaudeAndOpenCode(t *testing.T) {
	dir := t.TempDir()
	cfg := `{"enforce":{"claude":true,"opencode":true},"mcp":{"linear":{"command":"npx","args":["-y","linear-mcp"],"env":{"K":"${LINEAR_KEY}"}}}}`
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	// Claude dedicated file, generated + env-var preserved.
	mcpFile := filepath.Join(dir, ".mcp.json")
	if b, err := os.ReadFile(mcpFile); err != nil || !strings.Contains(string(b), "mcpServers") || !strings.Contains(string(b), "${LINEAR_KEY}") {
		t.Fatalf(".mcp.json should be generated with the env ref, got err=%v\n%s", err, b)
	}
	// OpenCode shared config, merged local server.
	if b, err := os.ReadFile(filepath.Join(dir, "opencode.json")); err != nil || !strings.Contains(string(b), `"mcp"`) || !strings.Contains(string(b), "linear") {
		t.Fatalf("opencode.json should carry the merged mcp server, got err=%v\n%s", err, b)
	}

	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass with MCP generated, got %v\n%s", err, out.String())
	}

	// Corrupt the OpenCode mcp entry → drift; --fix restores.
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(`{"mcp":{"linear":{"type":"local","command":["wrong"]}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("altered opencode mcp should drift, got code %d", code)
	}
	if !strings.Contains(out.String(), "opencode.json") {
		t.Fatalf("drift should name opencode.json, got: %s", out.String())
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", "--fix", dir}, &out, &errb); err != nil {
		t.Fatalf("check --fix: %v\n%s", err, out.String())
	}
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass after --fix, got %v", err)
	}

	// Removing the Claude dedicated file drifts too (it's a generated mirror).
	if err := os.Remove(mcpFile); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("missing .mcp.json should drift, got code %d", code)
	}
}

// The sync script reconstructs ALL dedicated gyroscope-owned files — not just
// persona mirrors — so a command file and the dedicated .mcp.json come back too
// (with the env-var ref intact, never a secret).
func TestSyncBakesCommandsAndMCP(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	dir := t.TempDir()
	cfg := `{"sync":true,"enforce":{"claude":true},"mcp":{"linear":{"command":"npx","args":["-y","linear-mcp"],"env":{"K":"${LINEAR_KEY}"}}}}`
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "docs", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "commands", "deploy.md"), []byte("Deploy $ARGUMENTS.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	cmdFile := filepath.Join(dir, ".claude", "commands", "deploy.md")
	mcpFile := filepath.Join(dir, ".mcp.json")
	wantCmd, err := os.ReadFile(cmdFile)
	if err != nil {
		t.Fatal(err)
	}
	wantMCP, err := os.ReadFile(mcpFile)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a fresh clone: dedicated generated files gitignored + absent.
	os.RemoveAll(filepath.Join(dir, ".claude"))
	os.Remove(mcpFile)

	c := exec.Command("sh", "gyroscope-sync.sh")
	c.Dir = dir
	if b, err := c.CombinedOutput(); err != nil {
		t.Fatalf("sync script failed: %v\n%s", err, b)
	}
	if got, err := os.ReadFile(cmdFile); err != nil || !bytes.Equal(got, wantCmd) {
		t.Fatalf("command file not reconstructed byte-for-byte (err=%v)", err)
	}
	got, err := os.ReadFile(mcpFile)
	if err != nil || !bytes.Equal(got, wantMCP) {
		t.Fatalf(".mcp.json not reconstructed byte-for-byte (err=%v)", err)
	}
	if !strings.Contains(string(got), "${LINEAR_KEY}") {
		t.Errorf("reconstructed .mcp.json should keep the env-var ref (no secret leak):\n%s", got)
	}
}

func TestCheckOnStateRequiresPersonaFile(t *testing.T) {
	dir := initAndFill(t)
	var out, errb bytes.Buffer
	if err := run([]string{"agents", "set", "on", dir}, &out, &errb); err != nil {
		t.Fatalf("agents set on: %v (%s)", err, errb.String())
	}
	out.Reset()
	errb.Reset()
	err := run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("on-state with no persona file should drift, got code %d (%v)", code, err)
	}
	if !strings.Contains(out.String(), "docs/agents/") {
		t.Fatalf("drift should name docs/agents/, got: %s", out.String())
	}
	// Add a persona file → conformant.
	if err := os.WriteFile(filepath.Join(dir, "docs", "agents", "code-reviewer.md"), []byte("# reviewer\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("on-state with a persona file should be conformant, got %v\n%s", err, out.String())
	}
}

func TestCheckFlagsMissingPersonasDirective(t *testing.T) {
	dir := initAndFill(t)
	hub, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	stripped := strings.ReplaceAll(string(hub), "spokes.personas", "SPOKES-REMOVED")
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(stripped), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err = run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("missing directive should drift, got code %d (%v)", code, err)
	}
	if !strings.Contains(out.String(), "personas directive") {
		t.Fatalf("drift should mention the personas directive, got: %s", out.String())
	}
}

func TestCheckFlagsMissingADRDirective(t *testing.T) {
	dir := initAndFill(t)
	hub, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	// Blank the directive heading inside the managed region so the standing text
	// no longer matches.
	stripped := strings.ReplaceAll(string(hub), "## Architecture decisions", "## Removed")
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(stripped), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err = run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("missing ADR directive should drift, got code %d (%v)", code, err)
	}
	if !strings.Contains(out.String(), "ADR directive") {
		t.Fatalf("drift should mention the ADR directive, got: %s", out.String())
	}
}

// A structural ADR defect (here, a numbering gap) is drift the ADR spoke polices
// beyond mere presence — the "keep them healthy" half of ADR enforcement.
func TestCheckFlagsADRNumberingGap(t *testing.T) {
	dir := initAndFill(t)
	// init scaffolds TEMPLATE.md + 0001; add a 0003 (skipping 0002) to open a gap.
	gap := "# ADR 0003: gap\n\n- **Status:** accepted\n- **Date:** 2026-01-01\n\n## Context\n\nx\n"
	if err := os.WriteFile(filepath.Join(dir, "docs", "adr", "0003-gap.md"), []byte(gap), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("ADR numbering gap should drift, got code %d (%v)", code, err)
	}
	if !(strings.Contains(out.String(), "docs/adr") && strings.Contains(out.String(), "0002")) {
		t.Fatalf("drift should name the missing ADR 0002, got: %s", out.String())
	}
}

func TestCheckIgnoresUserContentOutsideManagedRegion(t *testing.T) {
	dir := initAndFill(t)
	hub, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	// A user adds their own prose AFTER the managed region. gyroscope owns only
	// the managed block, so this must NOT be reported as drift, and any route-like
	// bullet the user writes outside the block must NOT count as a hub route.
	extra := string(hub) + "\n## My own section\n\n" +
		"- **Something the user cares about** → `docs/mine.md`.\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(extra), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("user content outside the managed region must not drift, got %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "conformant") {
		t.Fatalf("expected conformant, got: %s", out.String())
	}
}

func TestCheckFailsWhenManagedRegionMissing(t *testing.T) {
	dir := initAndFill(t)
	// Strip the managed markers entirely: gyroscope can no longer locate its
	// region, which is drift (the hub is no longer in the managed-block form).
	hub, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	stripped := strings.ReplaceAll(strings.ReplaceAll(string(hub),
		standard.ManagedOpen, ""), standard.ManagedClose, "")
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(stripped), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err = run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("a hub with no managed region should drift, got code %d (%v)", code, err)
	}
	if !strings.Contains(out.String(), "managed region") {
		t.Fatalf("drift should name the missing managed region, got: %s", out.String())
	}
}

func TestCheckFixConvergesMissingSpoke(t *testing.T) {
	dir := initAndFill(t)
	// Introduce drift: delete a ready-to-use (placeholder-free) spoke. --fix
	// re-creates it, and the repo is conformant again. (A placeholder-bearing
	// spoke would come back as a raw scaffold — structure only — because the
	// binary never fills placeholders; that is the skill's job.)
	if err := os.Remove(filepath.Join(dir, "CONTRIBUTING.md")); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("check --fix should converge and exit 0, got %v\n%s", err, out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "CONTRIBUTING.md")); err != nil {
		t.Fatalf("check --fix should have re-created the missing spoke: %v", err)
	}
	// A plain check afterward is conformant.
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("repo should be conformant after --fix, got %v\n%s", err, out.String())
	}
}

// check --fix converges the archive nudge too: it moves completed top-level tasks
// out of the injected TODO.md into DONE.md, so a later plain check is quiet. This is
// the mechanism that closes ADR 0009's convention-with-no-teeth gap — the nudge was
// the only finding --fix could not fix.
func TestCheckFixArchivesCompletedItems(t *testing.T) {
	dir := initAndFill(t)
	setTODO(t, dir, "# TODO\n\n## Next\n- [ ] keep me open\n"+
		"- [x] done a\n- [x] done b\n- [x] done c\n- [x] done d\n- [x] done e\n- [x] done f\n")

	var out, errb bytes.Buffer
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("check --fix should archive and exit 0, got %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "archived") {
		t.Fatalf("expected an 'archived N' line, got: %s", out.String())
	}

	todo, err := os.ReadFile(filepath.Join(dir, "TODO.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(todo), "- [x]") {
		t.Fatalf("TODO.md should have no completed items after --fix:\n%s", todo)
	}
	if !strings.Contains(string(todo), "keep me open") {
		t.Fatalf("--fix must keep open items in TODO.md:\n%s", todo)
	}

	done, err := os.ReadFile(filepath.Join(dir, "DONE.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []string{"done a", "done f"} {
		if !strings.Contains(string(done), item) {
			t.Fatalf("DONE.md should contain archived %q:\n%s", item, done)
		}
	}

	// A plain check afterward is conformant and no longer nudges.
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("repo should be conformant after archiving, got %v\n%s", err, out.String())
	}
	if strings.Contains(out.String(), "archive them to DONE.md") {
		t.Fatalf("nudge should be gone after --fix archived the items, got: %s", out.String())
	}
}

// A repo that adopted gyroscope before the contributor block existed has a
// CONTRIBUTING.md with no managed region. `check` flags the missing block as drift,
// and `check --fix` appends it (MergeManaged's markerless path) while preserving the
// user's own prose — the "maintained repos stay first-class for tool-less devs" path.
func TestCheckFixConvergesContributorBlock(t *testing.T) {
	dir := initAndFill(t)
	userDoc := "# Contributing\n\nOur own house rules. Keep them.\n"
	if err := os.WriteFile(filepath.Join(dir, "CONTRIBUTING.md"), []byte(userDoc), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("expected drift for the missing contributor block, got code %d\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "CONTRIBUTING.md") {
		t.Fatalf("drift should name CONTRIBUTING.md, got: %s", out.String())
	}

	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("check --fix should converge the contributor block, got %v\n%s", err, out.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "CONTRIBUTING.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "Our own house rules") {
		t.Fatalf("--fix must preserve the user's CONTRIBUTING prose:\n%s", got)
	}
	if !strings.Contains(string(got), "do not need gyroscope installed") {
		t.Fatalf("--fix must append the contributor block:\n%s", got)
	}
	// The whole block must land, not a marker-truncated prefix: assert its tail
	// (the zero-install command) and a well-formed closing marker are present.
	if !strings.Contains(string(got), "go run github.com/WagnerJust/gyroscope/cmd/gyroscope@latest check .") {
		t.Fatalf("--fix wrote a truncated contributor block (missing the zero-install tail):\n%s", got)
	}
	if _, ok := standard.ManagedRegion(got); !ok {
		t.Fatalf("--fix wrote a contributor block without a well-formed managed region:\n%s", got)
	}

	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("repo should be conformant after --fix, got %v\n%s", err, out.String())
	}
}

// A hand-edited contributor block (markers intact, inner text changed) is drift;
// check --fix swaps the managed region back to the standard.
func TestCheckFlagsAndFixesDriftedContributorBlock(t *testing.T) {
	dir := initAndFill(t)
	p := filepath.Join(dir, "CONTRIBUTING.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	drifted := strings.Replace(string(b), "You do not need gyroscope installed", "You TOTALLY need gyroscope", 1)
	if drifted == string(b) {
		t.Fatal("test setup failed to alter the managed region")
	}
	if err := os.WriteFile(p, []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("expected drift for the altered block, got code %d\n%s", code, out.String())
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("--fix should restore the block, got %v\n%s", err, out.String())
	}
	got, _ := os.ReadFile(p)
	if strings.Contains(string(got), "TOTALLY need") {
		t.Fatalf("--fix should have reverted the drifted managed region:\n%s", got)
	}
}

func TestCheckFixMergesManagedRegion(t *testing.T) {
	dir := initAndFill(t)
	// Drift the hub's managed region (blank it out) but keep user content around it.
	hub, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	region, ok := standard.ManagedRegion(hub)
	if !ok {
		t.Fatal("test hub should have a managed region")
	}
	drifted := strings.Replace(string(hub),
		standard.ManagedOpen+string(region)+standard.ManagedClose,
		standard.ManagedOpen+"\n"+standard.ManagedClose, 1)
	drifted += "\n## User addition\n\nkeep me\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("check --fix should merge the managed region, got %v\n%s", err, out.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "keep me") {
		t.Fatal("--fix must preserve user content outside the managed region")
	}
	if !strings.Contains(string(got), "docs/agents.md") {
		t.Fatal("--fix must re-inject the managed routes")
	}
}

func TestCheckFixRefusesConflict(t *testing.T) {
	dir := initAndFill(t)
	// A CONFLICT (foreign pointer with no managed region) is NOT auto-fixed — that
	// would clobber user content. --fix converges the safe subset and still reports
	// the conflict as drift (exit 1), so CI/dev see it needs --force.
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("my notes, no routing line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"check", dir, "--fix"}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("--fix with an unresolved conflict should still report drift, got code %d (%v)", code, err)
	}
	// The user's conflicting file is left untouched (not clobbered).
	if b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); string(b) != "my notes, no routing line\n" {
		t.Fatal("--fix must not clobber a conflicting file")
	}
}

func TestCheckPassesWithDisabledSpoke(t *testing.T) {
	dir := t.TempDir()
	// Context spoke off: neither CONTEXT.md nor its hub route should exist, and
	// their correct absence must NOT be reported as drift.
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"spokes":{"context":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)
	if _, err := os.Stat(filepath.Join(dir, "CONTEXT.md")); !os.IsNotExist(err) {
		t.Fatal("disabled context spoke should not have been written")
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass when a disabled spoke's route is correctly absent, got %v\nstdout: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "conformant") {
		t.Fatalf("expected a conformant confirmation, got: %s", out.String())
	}
}

// enforce.aiAttribution=false suppresses AI attribution via the Claude-native lever:
// init writes includeCoAuthoredBy=false, check flags a non-false value as drift, and
// --fix converges it back. gyroscope manages the key only under suppression.
func TestAIAttributionSuppression(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"claude":true,"aiAttribution":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	assertCoAuthored := func(want bool) {
		t.Helper()
		v, present, err := enforce.CoAuthoredBy(dir)
		if err != nil {
			t.Fatal(err)
		}
		if !present || v != want {
			t.Fatalf("includeCoAuthoredBy: present=%v value=%v, want present && %v", present, v, want)
		}
	}
	assertCoAuthored(false) // init suppressed it

	// Flip it back to true → check reports drift naming the key.
	if _, err := enforce.SetCoAuthoredBy(dir, true); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("expected drift when includeCoAuthoredBy=true under suppression, got code %d\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "includeCoAuthoredBy") {
		t.Fatalf("drift should name includeCoAuthoredBy, got: %s", out.String())
	}

	// --fix converges it back to false.
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("--fix should re-suppress attribution and exit 0, got %v\n%s", err, out.String())
	}
	assertCoAuthored(false)
}

// With the Codex adapter on and attribution suppressed, init writes
// commit_attribution="" into .codex/config.toml, check flags a non-empty value as
// drift, and --fix converges it back — Codex's native attribution kill-switch.
func TestAIAttributionSuppressionCodex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"codex":true,"aiAttribution":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	if val, present, err := enforce.CodexCommitAttribution(dir); err != nil || !present || val != "" {
		t.Fatalf("init should suppress: val=%q present=%v err=%v", val, present, err)
	}

	// A non-empty value → drift naming the key.
	if err := os.WriteFile(filepath.Join(dir, ".codex", "config.toml"), []byte(`commit_attribution = "Codex <noreply@openai.com>"`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("expected drift when commit_attribution is non-empty under suppression, got code %d\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "commit_attribution") {
		t.Fatalf("drift should name commit_attribution, got: %s", out.String())
	}

	// --fix converges it back to "".
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("--fix should re-suppress and exit 0, got %v\n%s", err, out.String())
	}
	if val, _, _ := enforce.CodexCommitAttribution(dir); val != "" {
		t.Fatalf("--fix should reset commit_attribution to empty, got %q", val)
	}
}

// The default (attribution on) never writes the key — gyroscope leaves Claude's own
// default in place and check does not police it.
func TestAIAttributionDefaultLeavesKeyUnmanaged(t *testing.T) {
	dir := initAndFill(t) // default config: aiAttribution on
	if _, present, err := enforce.CoAuthoredBy(dir); err != nil || present {
		t.Fatalf("default config must not write includeCoAuthoredBy (present=%v err=%v)", present, err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("attribution-on repo should be conformant, got %v\n%s", err, out.String())
	}
}

// Cross-harness half of the toggle: suppressing attribution injects a managed hub
// directive (so PI/Cursor/etc. honor it); check verifies it and --fix restores it.
func TestAIAttributionHubDirective(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"claude":true,"aiAttribution":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	fillPlaceholders(t, dir)

	hub, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(hub), "## AI attribution") || !strings.Contains(string(hub), "Co-Authored-By") {
		t.Fatalf("suppressed repo hub should carry the no-attribution directive:\n%s", hub)
	}

	// Alter the directive inside the managed region → drift; --fix restores it.
	region, ok := standard.ManagedRegion(hub)
	if !ok {
		t.Fatal("hub should have a managed region")
	}
	drifted := strings.Replace(string(hub), string(region),
		strings.Replace(string(region), "## AI attribution", "## Gone", 1), 1)
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := exitCodeOf(t, run([]string{"check", dir}, &out, &errb)); code != exitDrift {
		t.Fatalf("expected drift when the attribution directive is altered, got %d\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "attribution directive") {
		t.Fatalf("drift should name the attribution directive, got: %s", out.String())
	}
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("--fix should restore the directive and exit 0, got %v\n%s", err, out.String())
	}
	hub2, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(string(hub2), "## AI attribution") {
		t.Fatalf("--fix should have restored the directive:\n%s", hub2)
	}
}

package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

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

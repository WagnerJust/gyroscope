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

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/standard"
)

func TestInitDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write AGENTS.md")
	}
	s := out.String()
	if !strings.Contains(s, "dry-run") {
		t.Fatalf("expected dry-run notice, got: %s", s)
	}
	// The dry-run must surface the exact hook command that runs every session.
	if !strings.Contains(s, "cat AGENTS.md") {
		t.Fatalf("dry-run should show the SessionStart hook command, got: %s", s)
	}
	// Default config → local spoke on → init mutates .gitignore; surface it.
	if !strings.Contains(s, ".gitignore") {
		t.Fatalf("dry-run should note the .gitignore mutation, got: %s", s)
	}
}

func TestInitDryRunOmitsGitignoreWhenNoLocalWrites(t *testing.T) {
	dir := t.TempDir()
	// Both .local/-writing spokes off (local notes + local todo) → nothing lands
	// under .local/, so the dry-run must not advertise a .gitignore mutation.
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"spokes":{"local":false,"state":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	s := out.String()
	if strings.Contains(s, ".gitignore") {
		t.Fatalf("dry-run must not note .gitignore when nothing writes under .local/, got: %s", s)
	}
	if strings.Contains(s, ".local/local.md") || strings.Contains(s, ".local/todo.md") {
		t.Fatalf("dry-run hook must not include a disabled .local/ spoke, got: %s", s)
	}
	// Still writes nothing.
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write AGENTS.md")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write settings.json")
	}
}

func TestInitApplyWritesStandardPointerAndHook(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	for _, want := range []string{"AGENTS.md", "CLAUDE.md", "GEMINI.md", ".claude/settings.json", ".local/local.md"} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("apply should have written %s: %v", want, err)
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	if !strings.Contains(string(b), "SessionStart") || !strings.Contains(string(b), "cat AGENTS.md") {
		t.Fatalf("settings.json should contain the SessionStart hook command, got: %s", b)
	}
}

func TestInitApplyWritesCustomSpokeAndRoute(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"custom":[{"name":"Security","dest":"docs/security.md"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "security.md")); err != nil {
		t.Errorf("apply should have written the custom spoke docs/security.md: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, "docs/security.md") || !strings.Contains(got, "**Security**") {
		t.Fatalf("AGENTS.md should route to the custom spoke, got:\n%s", got)
	}
	if strings.Contains(got, "gyroscope:custom-routes") {
		t.Fatalf("AGENTS.md must not contain the raw marker, got:\n%s", got)
	}
}

func TestInitApplyConfigAwareHookOmitsDisabledSpoke(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"spokes":{"local":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	got := string(b)
	if strings.Contains(got, ".local/local.md") {
		t.Fatalf("hook must not cat a disabled spoke, got: %s", got)
	}
	if !strings.Contains(got, "AGENTS.md") || !strings.Contains(got, "docs/agents.md") {
		t.Fatalf("hook must cat the hub and enabled agents spoke, got: %s", got)
	}
}

func TestInitInjectsStateFilesInHookAndWritesThem(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, "TODO.md") || !strings.Contains(got, ".local/todo.md") {
		t.Fatalf("hook should cat the state files so a fresh session resumes, got: %s", got)
	}
	for _, p := range []string{"TODO.md", ".local/todo.md"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf("init should write the state file %s: %v", p, err)
		}
	}
}

// A single colliding pointer no longer aborts the whole run (that was the old
// all-or-nothing behavior). init --apply writes the safe subset around it and
// leaves only the conflicting file untouched — the conflict surfaces as drift.
func TestInitApplyWritesSafeSubsetAroundAConflict(t *testing.T) {
	dir := t.TempDir()
	// Pre-existing file that collides with one of the pointers, to prove the rest
	// of the standard still lands and only this file is skipped.
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err == nil {
		t.Fatal("a remaining conflict should surface as drift (exit 1)")
	}
	// The safe subset lands regardless of the collision.
	for _, p := range []string{"AGENTS.md", "CLAUDE.md", ".claude/settings.json"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf("safe subset should land despite a conflict: %s missing (%v)", p, err)
		}
	}
	// The conflicting file is untouched, and reported skipped.
	if b, _ := os.ReadFile(filepath.Join(dir, "GEMINI.md")); string(b) != "mine" {
		t.Fatal("existing conflicting file must be untouched")
	}
	if !strings.Contains(out.String(), "GEMINI.md") {
		t.Fatalf("the skipped conflict should be reported, got:\n%s", out.String())
	}
}

func TestInitApplyIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("first apply: %v (%s)", err, errb.String())
	}
	before, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	// A second --apply with everything already current must SUCCEED (no --force
	// needed) and change nothing — init is merge-safe and idempotent (D3).
	out.Reset()
	errb.Reset()
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("re-apply on a current repo should succeed, got %v (%s)", err, errb.String())
	}
	after, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("re-apply must not rewrite an already-current hub")
	}
}

func TestInitApplyMergesManagedRegionIntoExistingHub(t *testing.T) {
	dir := t.TempDir()
	// The user already has an AGENTS.md with their own prose and only an empty
	// managed region. init --apply (no --force) must inject the managed content
	// AND preserve the user's surrounding prose — a MERGE, not a refuse.
	userHub := "# AGENTS.md\n\n" +
		"Welcome to my repo — read the house rules below.\n\n" +
		standard.ManagedOpen + "\n" + standard.ManagedClose + "\n\n" +
		"## House rules (mine)\n\nBe kind. Ship small.\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(userHub), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("merge apply should succeed without --force, got %v (%s)", err, errb.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	// User prose preserved.
	if !strings.Contains(s, "Welcome to my repo") || !strings.Contains(s, "## House rules (mine)") {
		t.Fatalf("merge must preserve the user's content, got:\n%s", s)
	}
	// Managed content injected (a real route now lives inside the region).
	if !strings.Contains(s, "**Build, test, conventions** → `docs/agents.md`.") {
		t.Fatalf("merge must inject the managed routes, got:\n%s", s)
	}
	// Other planned files were created too (NEW subset applied).
	if _, err := os.Stat(filepath.Join(dir, "CONTEXT.md")); err != nil {
		t.Errorf("merge apply should also create the NEW files: %v", err)
	}
	// And the repo is conformant after the merge.
	fillPlaceholders(t, dir)
	out.Reset()
	errb.Reset()
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("repo should be conformant after a merge apply, got %v\n%s", err, out.String())
	}
}

// Regression (defect 1, apply half): a markerless hand-written hub is merged in
// place on --apply (no --force) — the managed region is injected AND every byte
// of the user's original hub survives. Pairs with TestClassifyMergeOnMarkerlessHub.
func TestInitApplyMergesMarkerlessHub(t *testing.T) {
	dir := t.TempDir()
	// A hand-written hub with sentinel content and no gyroscope markers.
	userHub := "# notwhoop — agent hub\n\n" +
		"Before touching the band, read this.\n\n" +
		"## Milestones\n\n1. openwhoop → local SQLite buffer\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(userHub), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("markerless hub merge should succeed without --force, got %v (%s)", err, errb.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	// The managed region was injected.
	if !strings.Contains(s, standard.ManagedOpen) || !strings.Contains(s, standard.ManagedClose) {
		t.Fatalf("apply must inject the managed region into a markerless hub, got:\n%s", s)
	}
	if !strings.Contains(s, "**Build, test, conventions** → `docs/agents.md`.") {
		t.Fatalf("apply must inject the managed routes, got:\n%s", s)
	}
	// Sentinels from the original hand-written hub survive — nothing clobbered.
	if !strings.Contains(s, "Before touching the band, read this.") {
		t.Fatalf("original hub content must survive the merge, got:\n%s", s)
	}
	if !strings.Contains(s, "openwhoop → local SQLite buffer") {
		t.Fatalf("original milestone content must survive the merge, got:\n%s", s)
	}
}

// Regression (defect 2): with a CONFLICT present, init --apply (no --force) must
// apply the NEW + MERGE subset and SKIP only the conflict — not refuse the whole
// run. D3's spec: "Only a true CONFLICT needs --force."
func TestInitApplySkipsConflictsAndWritesTheSafeSubset(t *testing.T) {
	dir := t.TempDir()
	// One CONFLICT (a foreign pointer with its own content, no managed region).
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("my own claude notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"init", dir, "--apply"}, &out, &errb)
	// The safe subset landing while a conflict remains is drift (exit 1), not a
	// can't-run: the repo is not yet fully conformant, --force is still needed.
	if err == nil {
		t.Fatal("a remaining conflict should surface as drift (exit 1), got success")
	}
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("skipped conflict should exit %d (drift), got %d: %v", exitDrift, code, err)
	}
	// The NEW files landed even though CLAUDE.md conflicts.
	for _, p := range []string{"CONTEXT.md", "GEMINI.md", ".cursorrules", "docs/agents.md", "TODO.md"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf("NEW file %s should have landed alongside a skipped conflict: %v", p, err)
		}
	}
	// The conflicting file is untouched.
	if b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); string(b) != "my own claude notes\n" {
		t.Fatal("the conflicting file must be left untouched")
	}
	// The skip is reported clearly, naming the file and pointing at --force.
	s := out.String()
	if !strings.Contains(s, "skipped") || !strings.Contains(s, "--force") || !strings.Contains(s, "CLAUDE.md") {
		t.Fatalf("apply should report the skipped conflict (use --force): CLAUDE.md, got:\n%s", s)
	}
}

func TestInitApplyRefusesConflictWithoutForce(t *testing.T) {
	dir := t.TempDir()
	// A foreign pointer file with no managed region is a CONFLICT: init --apply
	// skips it (writing the safe subset) but leaves it untouched, and surfaces the
	// remaining conflict as drift.
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("my own claude notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"init", dir, "--apply"}, &out, &errb)
	if err == nil {
		t.Fatal("a CONFLICT should surface as drift without --force")
	}
	if !strings.Contains(out.String(), "CLAUDE.md") {
		t.Fatalf("the skip report should name the conflicting file, got:\n%s", out.String())
	}
	// The safe subset still lands (no longer all-or-nothing): the hub is written.
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatalf("the safe subset should land while a conflict is skipped: %v", err)
	}
	// User's file untouched.
	if b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); string(b) != "my own claude notes\n" {
		t.Fatal("conflicting file must be left untouched")
	}
	// With --force the conflict is overwritten.
	out.Reset()
	errb.Reset()
	if err := run([]string{"init", dir, "--apply", "--force"}, &out, &errb); err != nil {
		t.Fatalf("--force should resolve the conflict, got %v (%s)", err, errb.String())
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); !strings.Contains(string(b), "read AGENTS.md") {
		t.Fatal("--force should have written the canonical routing line")
	}
}

func TestInitInstallsPIExtensionWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"pi":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".pi", "extensions", "gyroscope-context.ts")); err != nil {
		t.Fatalf("PI extension should be written when enabled: %v", err)
	}
}

func TestInitDryRunListsPIExtensionWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"pi":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir}, &out, &errb); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), ".pi/extensions/gyroscope-context.ts") {
		t.Fatalf("dry-run should list the PI extension, got:\n%s", out.String())
	}
}

func TestInitSkipsPIExtensionByDefault(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".pi")); !os.IsNotExist(err) {
		t.Fatal("PI is opt-in; .pi/ should not be created by default")
	}
}

// DONE.md is enforced (Plan/check) and hub-routed, but must NEVER be catted by
// the SessionStart hook — the whole point of the TODO/DONE split is to keep the
// completed archive out of the per-session injection (ADR 0009).
func TestHookPathsNeverCatsDone(t *testing.T) {
	if contains(hookPathsFor(config.Default()), "DONE.md") {
		t.Fatal("hook must never cat DONE.md — the archive is routed, not injected")
	}
}

func TestHookPathsCatGyroscopeJSONWhenPersonasEnabled(t *testing.T) {
	cfg := config.Default() // personas unknown → enabled
	got := hookPathsFor(cfg)
	if !contains(got, "gyroscope.json") {
		t.Fatalf("hook should cat gyroscope.json when personas enabled: %v", got)
	}

	cfg.Spokes.Personas = config.PersonaOff
	got = hookPathsFor(cfg)
	if contains(got, "gyroscope.json") {
		t.Fatalf("hook should not cat gyroscope.json when personas off: %v", got)
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

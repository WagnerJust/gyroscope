package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WagnerJust/gyroscope/internal/config"
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

func TestInitApplyIsAllOrNothing(t *testing.T) {
	dir := t.TempDir()
	// Pre-existing file that collides with one of the pointers (not the first standard file),
	// to prove NOTHING is written when any target collides.
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err == nil {
		t.Fatal("expected refusal when a target already exists")
	}
	// Nothing else should have been written — AGENTS.md / CLAUDE.md / .claude must be absent.
	for _, p := range []string{"AGENTS.md", "CLAUDE.md", ".claude/settings.json"} {
		if _, err := os.Stat(filepath.Join(dir, p)); !os.IsNotExist(err) {
			t.Errorf("all-or-nothing violated: %s exists", p)
		}
	}
	// The pre-existing file is untouched.
	if b, _ := os.ReadFile(filepath.Join(dir, "GEMINI.md")); string(b) != "mine" {
		t.Fatal("existing file must be untouched")
	}
}

func TestInitApplyRefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("first apply: %v (%s)", err, errb.String())
	}
	// A second --apply without --force must refuse to clobber the standard
	// (protecting content the companion skill may have filled in).
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err == nil {
		t.Fatal("second apply without --force should refuse to overwrite")
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

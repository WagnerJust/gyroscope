package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if !strings.Contains(out.String(), "dry-run") {
		t.Fatalf("expected dry-run notice, got: %s", out.String())
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

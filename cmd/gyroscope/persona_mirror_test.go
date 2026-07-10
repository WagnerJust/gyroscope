package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writePersona is a test helper that writes a persona file under docs/agents/.
func writePersona(t *testing.T, dir, name, content string) {
	t.Helper()
	dest := filepath.Join(dir, "docs", "agents", name)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const samplePersona = `---
name: verifier
description: Verify things.
tools: Read, Bash
---

You are the verifier persona for this repo.
`

// personasOnConfig turns personas on and keeps the Claude adapter on (the gate).
const personasOnConfig = `{"spokes":{"personas":"on"}}`

func TestInitApplyMirrorsPersonasWhenGated(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(personasOnConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	writePersona(t, dir, "verifier.md", samplePersona)

	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}

	mirror := filepath.Join(dir, ".claude", "agents", "verifier.md")
	got, err := os.ReadFile(mirror)
	if err != nil {
		t.Fatalf("expected .claude/agents/verifier.md to be written: %v", err)
	}
	if string(got) != samplePersona {
		t.Errorf(".claude/agents/verifier.md not byte-equal to source")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "README.md")); !os.IsNotExist(err) {
		t.Errorf("README.md must not be mirrored")
	}
}

func TestInitDryRunListsPersonaMirrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(personasOnConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	writePersona(t, dir, "verifier.md", samplePersona)

	var out, errb bytes.Buffer
	if err := run([]string{"init", dir}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	s := out.String()
	if !strings.Contains(s, ".claude/agents/verifier.md") {
		t.Fatalf("dry-run should list the persona mirror it would write, got:\n%s", s)
	}
	// Dry-run writes nothing.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "verifier.md")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write the persona mirror")
	}
}

// The gate: personas must be `on` AND the Claude adapter enabled. With personas
// off (or unknown/skipped), or the Claude adapter disabled, no mirror is written.
func TestInitApplyNoMirrorWhenPersonasNotOn(t *testing.T) {
	dir := t.TempDir()
	// Personas default is `unknown` (not `on`) — write a persona file anyway.
	writePersona(t, dir, "verifier.md", samplePersona)

	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents")); !os.IsNotExist(err) {
		t.Errorf(".claude/agents/ must not be created when personas are not `on`")
	}
}

func TestInitApplyNoMirrorWhenClaudeDisabled(t *testing.T) {
	dir := t.TempDir()
	// Personas on, but the Claude enforcement adapter is off → gate closed.
	cfg := `{"spokes":{"personas":"on"},"enforce":{"claude":false}}`
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	writePersona(t, dir, "verifier.md", samplePersona)

	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("run: %v (%s)", err, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents")); !os.IsNotExist(err) {
		t.Errorf(".claude/agents/ must not be created when the Claude adapter is disabled")
	}
}

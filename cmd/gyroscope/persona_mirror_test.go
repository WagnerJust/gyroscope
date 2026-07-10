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

// initGatedWithPersona applies the standard into a fresh repo, fills the
// scaffolds, sets personas on, writes a persona, and mirrors it — the conformant
// registered baseline check should accept.
func initGatedWithPersona(t *testing.T) string {
	t.Helper()
	dir := initAndFill(t)
	var out, errb bytes.Buffer
	if err := run([]string{"agents", "set", "on", dir}, &out, &errb); err != nil {
		t.Fatalf("agents set on: %v (%s)", err, errb.String())
	}
	writePersona(t, dir, "verifier.md", samplePersona)
	out.Reset()
	errb.Reset()
	// init --apply again to mirror the just-added persona. The filled scaffolds now
	// differ from the raw templates, so this second apply reports drift (exit 1) on
	// those unrelated conflicts — but persona registration runs independent of doc
	// convergence, so the mirror is written regardless. Tolerate that drift here.
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		if code := exitCodeOf(t, err); code != exitDrift {
			t.Fatalf("init --apply (mirror): %v (%s)", err, errb.String())
		}
	}
	return dir
}

func TestCheckConformantWhenPersonaMirrored(t *testing.T) {
	dir := initGatedWithPersona(t)
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir}, &out, &errb); err != nil {
		t.Fatalf("check should pass when the persona is mirrored, got %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "conformant") {
		t.Fatalf("expected conformant, got: %s", out.String())
	}
}

func TestCheckDriftsWhenPersonaMirrorMissing(t *testing.T) {
	dir := initGatedWithPersona(t)
	// Remove the registered mirror → the persona is no longer dispatchable.
	if err := os.Remove(filepath.Join(dir, ".claude", "agents", "verifier.md")); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("missing mirror should drift, got code %d (%v)", code, err)
	}
	if !strings.Contains(out.String(), ".claude/agents/verifier.md") {
		t.Fatalf("drift should name the missing mirror, got: %s", out.String())
	}
}

func TestCheckDriftsWhenPersonaMirrorDiffers(t *testing.T) {
	dir := initGatedWithPersona(t)
	// Edit the registered mirror so it no longer matches the canonical source.
	if err := os.WriteFile(filepath.Join(dir, ".claude", "agents", "verifier.md"), []byte("stale edited mirror\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	err := run([]string{"check", dir}, &out, &errb)
	if code := exitCodeOf(t, err); code != exitDrift {
		t.Fatalf("differing mirror should drift, got code %d (%v)", code, err)
	}
	if !strings.Contains(out.String(), ".claude/agents/verifier.md") {
		t.Fatalf("drift should name the differing mirror, got: %s", out.String())
	}
}

func TestCheckFixReMirrorsPersona(t *testing.T) {
	dir := initGatedWithPersona(t)
	mirror := filepath.Join(dir, ".claude", "agents", "verifier.md")
	// Drift the mirror, then --fix should re-mirror it to convergence.
	if err := os.WriteFile(mirror, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"check", dir, "--fix"}, &out, &errb); err != nil {
		t.Fatalf("check --fix should converge, got %v\n%s", err, out.String())
	}
	got, err := os.ReadFile(mirror)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != samplePersona {
		t.Errorf("--fix did not re-mirror the persona; got %q", got)
	}
}

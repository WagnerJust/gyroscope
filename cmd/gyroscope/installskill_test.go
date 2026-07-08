package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bytes"
)

func TestInstallSkillWritesEmbeddedSkill(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if err := run([]string{"install-skill", dir, "--apply"}, &out, &out); err != nil {
		t.Fatalf("run: %v (%s)", err, out.String())
	}
	dest := filepath.Join(dir, "gyroscope", "SKILL.md")
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("skill not written: %v", err)
	}
	if !strings.Contains(string(b), "name: gyroscope") {
		t.Fatal("written skill missing frontmatter")
	}
}

func TestInstallSkillRejectsTooManyArgs(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"install-skill", "a", "b", "--apply"}, &out, &out); err == nil {
		t.Fatal("expected error for too many args, got nil")
	}
}

func TestInstallSkillWarnsWhenBinaryNotOnPath(t *testing.T) {
	// Stub the PATH lookup to simulate `gyroscope` being absent, so the skill's
	// step-2 shell-out would fail. install-skill must warn explicitly with install
	// instructions rather than let that surface as a silent step-2 crash later.
	orig := lookBinary
	lookBinary = func(string) (string, error) { return "", os.ErrNotExist }
	defer func() { lookBinary = orig }()

	dir := t.TempDir()
	var out bytes.Buffer
	if err := run([]string{"install-skill", dir, "--apply"}, &out, &out); err != nil {
		t.Fatalf("install-skill should still succeed (skill written) even if the binary is absent: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "not on your PATH") {
		t.Fatalf("expected a not-on-PATH warning, got:\n%s", s)
	}
	if !strings.Contains(s, "go install") {
		t.Fatalf("warning should include an install instruction (go install), got:\n%s", s)
	}
	// The skill itself is still installed.
	if _, err := os.Stat(filepath.Join(dir, "gyroscope", "SKILL.md")); err != nil {
		t.Fatalf("skill should still be written: %v", err)
	}
}

func TestInstallSkillQuietWhenBinaryOnPath(t *testing.T) {
	orig := lookBinary
	lookBinary = func(string) (string, error) { return "/usr/local/bin/gyroscope", nil }
	defer func() { lookBinary = orig }()

	dir := t.TempDir()
	var out bytes.Buffer
	if err := run([]string{"install-skill", dir, "--apply"}, &out, &out); err != nil {
		t.Fatalf("run: %v (%s)", err, out.String())
	}
	if strings.Contains(out.String(), "not on your PATH") {
		t.Fatalf("no warning expected when the binary resolves, got:\n%s", out.String())
	}
}

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
	if err := run([]string{"install-skill", "--dir", dir, "--apply"}, &out, &out); err != nil {
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

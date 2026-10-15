package main

import (
	"path/filepath"
	"testing"
)

func TestDefaultSkillsDirHonoursConfigDir(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", "/cfg/personal")
	got, err := defaultSkillsDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/cfg/personal", "skills"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDefaultSkillsDirFallsBackToHome(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("HOME", "/home/x")
	got, err := defaultSkillsDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/home/x", ".claude", "skills"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

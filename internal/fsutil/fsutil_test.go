package fsutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteGuardedCreatesParentDirsAndWritesBytes(t *testing.T) {
	dir := t.TempDir()
	rel := filepath.Join("docs", "adr", "0001.md")
	content := []byte("hello guarded write\n")
	if err := WriteGuarded(dir, rel, content, false); err != nil {
		t.Fatalf("WriteGuarded: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("content = %q, want %q", got, content)
	}
}

func TestWriteGuardedRefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	rel := "CLAUDE.md"
	if err := os.WriteFile(filepath.Join(dir, rel), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := WriteGuarded(dir, rel, []byte("new"), false)
	if err == nil {
		t.Fatal("expected refuse-to-overwrite error")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite") || !strings.Contains(err.Error(), rel) {
		t.Fatalf("error = %q, want it to mention refusing to overwrite and %q", err, rel)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, rel)); string(b) != "mine" {
		t.Fatalf("existing file must be untouched, got %q", b)
	}
}

func TestWriteGuardedForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	rel := "CLAUDE.md"
	if err := os.WriteFile(filepath.Join(dir, rel), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := []byte("replaced\n")
	if err := WriteGuarded(dir, rel, want, true); err != nil {
		t.Fatalf("force WriteGuarded: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

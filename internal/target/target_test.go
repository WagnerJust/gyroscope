package target

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWritePointerCreatesRoutingLine(t *testing.T) {
	dir := t.TempDir()
	tgt, ok := ByID("claude")
	if !ok {
		t.Fatal("claude target missing")
	}
	if err := WritePointer(dir, tgt, false); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(b) != RoutingLine {
		t.Fatalf("pointer body = %q, want %q", b, RoutingLine)
	}
}

func TestWritePointerRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("mine"), 0o644)
	tgt, _ := ByID("claude")
	if err := WritePointer(dir, tgt, false); err == nil {
		t.Fatal("expected refuse-to-overwrite")
	}
}

func TestWritePointerForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	tgt, _ := ByID("claude")
	if err := WritePointer(dir, tgt, true); err != nil {
		t.Fatalf("force write should succeed over existing file: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(b) != RoutingLine {
		t.Fatalf("force-written pointer = %q, want %q", b, RoutingLine)
	}
}

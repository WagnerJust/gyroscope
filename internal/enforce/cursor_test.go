package enforce

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCursorApplyWritesAlwaysApplyRule(t *testing.T) {
	dir := t.TempDir()
	c := Cursor{}
	changed, err := c.Apply(dir, Paths{})
	if err != nil || !changed {
		t.Fatalf("first Apply: changed=%v err=%v", changed, err)
	}
	b, err := os.ReadFile(filepath.Join(dir, cursorRulePath))
	if err != nil {
		t.Fatalf("rule not written: %v", err)
	}
	if !strings.Contains(string(b), "alwaysApply: true") {
		t.Errorf("rule missing alwaysApply: true:\n%s", b)
	}
	if !strings.Contains(string(b), "read AGENTS.md and follow its routes") {
		t.Errorf("rule missing the hub directive:\n%s", b)
	}
}

func TestCursorApplyIsIdempotentAndVerifies(t *testing.T) {
	dir := t.TempDir()
	c := Cursor{}
	if _, err := c.Apply(dir, Paths{}); err != nil {
		t.Fatal(err)
	}
	changed, err := c.Apply(dir, Paths{})
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("second Apply should be a no-op (already current)")
	}
	ok, err := c.Verify(dir, Paths{})
	if err != nil || !ok {
		t.Fatalf("Verify after Apply: ok=%v err=%v", ok, err)
	}
}

func TestCursorVerifyFalseWhenMissingOrAltered(t *testing.T) {
	dir := t.TempDir()
	c := Cursor{}
	if ok, err := c.Verify(dir, Paths{}); err != nil || ok {
		t.Fatalf("missing rule should verify false: ok=%v err=%v", ok, err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cursorRulePath), []byte("---\nalwaysApply: false\n---\nedited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ok, _ := c.Verify(dir, Paths{}); ok {
		t.Error("altered rule should verify false")
	}
}

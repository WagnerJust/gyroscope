package enforce

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readOpenCode(t *testing.T, dir string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, opencodeConfigPath))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("parse opencode.json: %v", err)
	}
	return m
}

func TestOpenCodeInjectPathsDropsHubAndLocal(t *testing.T) {
	got := opencodeInjectPaths([]string{"AGENTS.md", "docs/agents.md", "TODO.md", ".local/todo.md", "gyroscope.json"})
	want := []string{"docs/agents.md", "TODO.md", "gyroscope.json"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestOpenCodeApplyMergesPreservingOtherKeys(t *testing.T) {
	dir := t.TempDir()
	// Pre-existing opencode.json with the user's own key + instruction entry.
	orig := `{"model":"anthropic/x","instructions":["MYRULES.md"]}`
	if err := os.WriteFile(filepath.Join(dir, opencodeConfigPath), []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	a := OpenCode{}
	paths := []string{"AGENTS.md", "TODO.md", "docs/agents.md"}
	changed, err := a.Apply(dir, paths)
	if err != nil || !changed {
		t.Fatalf("Apply: changed=%v err=%v", changed, err)
	}
	m := readOpenCode(t, dir)
	if m["model"] != "anthropic/x" {
		t.Errorf("user key not preserved: %v", m["model"])
	}
	instr, _ := m["instructions"].([]any)
	set := map[string]bool{}
	for _, e := range instr {
		set[e.(string)] = true
	}
	for _, want := range []string{"MYRULES.md", "TODO.md", "docs/agents.md"} {
		if !set[want] {
			t.Errorf("instructions missing %q: %v", want, instr)
		}
	}
	if set["AGENTS.md"] {
		t.Error("AGENTS.md must not be injected (read natively)")
	}

	// Idempotent: a second Apply is a no-op, and Verify passes.
	changed, err = a.Apply(dir, paths)
	if err != nil || changed {
		t.Fatalf("second Apply should be no-op: changed=%v err=%v", changed, err)
	}
	if ok, err := a.Verify(dir, paths); err != nil || !ok {
		t.Fatalf("Verify after Apply: ok=%v err=%v", ok, err)
	}
}

func TestOpenCodeVerifyFalseWhenMissing(t *testing.T) {
	dir := t.TempDir()
	a := OpenCode{}
	if ok, err := a.Verify(dir, []string{"TODO.md"}); err != nil || ok {
		t.Fatalf("absent opencode.json should verify false: ok=%v err=%v", ok, err)
	}
}

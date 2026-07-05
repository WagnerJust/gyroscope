package enforce

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeInstallAddsHookThenIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	c := Claude{}

	changed, err := c.Install(dir)
	if err != nil || !changed {
		t.Fatalf("first install: changed=%v err=%v", changed, err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("settings not valid json: %v", err)
	}
	if !strings.Contains(string(b), SessionStartCommand) {
		t.Fatalf("written settings should contain the hook command, got: %s", b)
	}

	changed, err = c.Install(dir)
	if err != nil || changed {
		t.Fatalf("second install should be a no-op: changed=%v err=%v", changed, err)
	}
}

func TestClaudeInstallPreservesExistingSettings(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude", "settings.json"),
		[]byte(`{"model":"opus","hooks":{"Stop":[{"hooks":[]}]}}`), 0o644)

	if _, err := (Claude{}).Install(dir); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	var m map[string]any
	json.Unmarshal(b, &m)
	if m["model"] != "opus" {
		t.Fatal("existing top-level setting was clobbered")
	}
	hooks := m["hooks"].(map[string]any)
	if _, ok := hooks["Stop"]; !ok {
		t.Fatal("existing Stop hook was clobbered")
	}
	if _, ok := hooks["SessionStart"]; !ok {
		t.Fatal("SessionStart hook not added")
	}
}

func TestClaudeInstallPreservesOtherToolsSessionStart(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	const other = "echo other-tool"
	seed := `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"echo other-tool"}]}]}}`
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := (Claude{}).Install(dir)
	if err != nil || !changed {
		t.Fatalf("install onto a populated SessionStart: changed=%v err=%v", changed, err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	got := string(b)
	if !strings.Contains(got, other) {
		t.Fatalf("another tool's SessionStart hook was dropped, got: %s", got)
	}
	if !strings.Contains(got, SessionStartCommand) {
		t.Fatalf("gyroscope hook not appended, got: %s", got)
	}
}

func TestClaudeInstallRejectsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	path := filepath.Join(dir, ".claude", "settings.json")
	const corrupt = `{"model":"opus"` // truncated — not valid JSON
	if err := os.WriteFile(path, []byte(corrupt), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := (Claude{}).Install(dir); err == nil {
		t.Fatal("expected an error on malformed settings.json")
	}
	if b, _ := os.ReadFile(path); string(b) != corrupt {
		t.Fatalf("malformed file must be left untouched, got: %s", b)
	}
}

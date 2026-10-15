package target

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WagnerJust/gyroscope/internal/config"
)

func TestRegistryHasExpectedTargets(t *testing.T) {
	want := map[string]string{
		"claude":   "CLAUDE.md",
		"gemini":   "GEMINI.md",
		"cursor":   ".cursorrules",
		"windsurf": ".windsurfrules",
		"cline":    ".clinerules",
		"copilot":  ".github/copilot-instructions.md",
		"zed":      ".rules",
	}
	for id, path := range want {
		tgt, ok := ByID(id)
		if !ok {
			t.Fatalf("target %q missing from registry", id)
		}
		if tgt.Path != path {
			t.Fatalf("target %q path = %q, want %q", id, tgt.Path, path)
		}
	}
	if n := len(All()); n != len(want) {
		t.Fatalf("registry has %d targets, want %d", n, len(want))
	}
}

const claudeBodyAgentsOn = "<!-- gyroscope: this file imports the hub so Claude Code loads it in full at launch. Edit AGENTS.md, not this file. -->\n" +
	"@AGENTS.md\n" +
	"@docs/agents.md\n"

func TestClaudeBodyImportsTheHub(t *testing.T) {
	tgt, _ := ByID("claude")
	cfg := config.Default()
	if got := string(tgt.Body(cfg)); got != claudeBodyAgentsOn {
		t.Fatalf("claude body (agents on):\n got %q\nwant %q", got, claudeBodyAgentsOn)
	}
	cfg.Spokes.Agents = false
	want := "<!-- gyroscope: this file imports the hub so Claude Code loads it in full at launch. Edit AGENTS.md, not this file. -->\n" +
		"@AGENTS.md\n"
	if got := string(tgt.Body(cfg)); got != want {
		t.Fatalf("claude body (agents off):\n got %q\nwant %q", got, want)
	}
}

func TestOtherTargetsKeepTheRoutingLine(t *testing.T) {
	for _, tgt := range All() {
		if tgt.ID == "claude" {
			continue
		}
		if got := string(tgt.Body(config.Default())); got != RoutingLine {
			t.Errorf("%s body = %q, want the routing line", tgt.ID, got)
		}
	}
}

func TestIsLegacyBody(t *testing.T) {
	if !IsLegacyBody([]byte(RoutingLine)) {
		t.Fatal("the routing line is the legacy body")
	}
	if IsLegacyBody([]byte(RoutingLine+"my notes\n")) || IsLegacyBody([]byte(claudeBodyAgentsOn)) {
		t.Fatal("only a byte-equal routing line is legacy")
	}
}

func TestWritePointerWritesTheBody(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	for _, id := range []string{"claude", "gemini"} {
		tgt, ok := ByID(id)
		if !ok {
			t.Fatalf("%s target missing", id)
		}
		if err := WritePointer(dir, tgt, cfg, false); err != nil {
			t.Fatal(err)
		}
		b, _ := os.ReadFile(filepath.Join(dir, tgt.Path))
		if string(b) != string(tgt.Body(cfg)) {
			t.Fatalf("%s pointer = %q, want %q", id, b, tgt.Body(cfg))
		}
	}
}

func TestWritePointerRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("mine"), 0o644)
	tgt, _ := ByID("claude")
	if err := WritePointer(dir, tgt, config.Default(), false); err == nil {
		t.Fatal("expected refuse-to-overwrite")
	}
}

func TestWritePointerForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	tgt, _ := ByID("gemini")
	if err := WritePointer(dir, tgt, config.Default(), true); err != nil {
		t.Fatalf("force write should succeed over existing file: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "GEMINI.md"))
	if string(b) != RoutingLine {
		t.Fatalf("force-written pointer = %q, want %q", b, RoutingLine)
	}
}

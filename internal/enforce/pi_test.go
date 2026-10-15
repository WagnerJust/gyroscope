package enforce

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPISatisfiesAdapter(t *testing.T) {
	var _ Adapter = PI{}
	if (PI{}).ID() != "pi" {
		t.Fatalf("PI ID should be pi, got %q", (PI{}).ID())
	}
}

func TestPIApplyWritesExtensionExcludingHub(t *testing.T) {
	dir := t.TempDir()
	paths := Paths{Rules: []string{"AGENTS.md", "docs/agents.md"}, State: []string{"gyroscope.json", "TODO.md"}}

	changed, err := (PI{}).Apply(dir, paths)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("first Apply should write the extension")
	}
	b, err := os.ReadFile(filepath.Join(dir, ".pi", "extensions", "gyroscope-context.ts"))
	if err != nil {
		t.Fatalf("extension should exist: %v", err)
	}
	src := string(b)
	if strings.Contains(src, `"AGENTS.md"`) {
		t.Fatal("PI injects the non-hub spokes only; AGENTS.md is read natively and must be excluded")
	}
	// Rules (minus the hub) first, then state, in order.
	if !strings.Contains(src, `const PATHS: string[] = ["docs/agents.md","gyroscope.json","TODO.md"];`) {
		t.Fatalf("extension PATHS should be rules-minus-hub then state, got:\n%s", src)
	}
	if !strings.Contains(src, "session_start") || !strings.Contains(src, "sendMessage") {
		t.Fatal("extension should inject on session_start via sendMessage")
	}
}

func TestPIApplyIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	paths := Paths{Rules: []string{"AGENTS.md"}, State: []string{"TODO.md"}}
	if _, err := (PI{}).Apply(dir, paths); err != nil {
		t.Fatal(err)
	}
	changed, err := (PI{}).Apply(dir, paths)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second Apply with identical paths should be idempotent")
	}
}

func TestPIVerifyRoundTrips(t *testing.T) {
	dir := t.TempDir()
	paths := Paths{Rules: []string{"AGENTS.md"}, State: []string{"TODO.md"}}

	ok, err := (PI{}).Verify(dir, paths)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Verify should be false before Apply")
	}
	if _, err := (PI{}).Apply(dir, paths); err != nil {
		t.Fatal(err)
	}
	ok, err = (PI{}).Verify(dir, paths)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Verify should be true after Apply")
	}
	// Alter the file → drift.
	dest := filepath.Join(dir, ".pi", "extensions", "gyroscope-context.ts")
	if err := os.WriteFile(dest, []byte("// tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err = (PI{}).Verify(dir, paths)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Verify should be false after the extension is altered")
	}
}

func TestPIPlanLineNamesTheExtension(t *testing.T) {
	line := (PI{}).PlanLine(Paths{Rules: []string{"AGENTS.md"}, State: []string{"TODO.md"}})
	if !strings.Contains(line, ".pi/extensions/gyroscope-context.ts") {
		t.Fatalf("plan line should name the extension, got %q", line)
	}
}

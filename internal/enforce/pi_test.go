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
	paths := []string{"AGENTS.md", "docs/agents.md", "TODO.md", "gyroscope.json"}

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
	for _, want := range []string{`"docs/agents.md"`, `"TODO.md"`, `"gyroscope.json"`} {
		if !strings.Contains(src, want) {
			t.Fatalf("extension PATHS should include %s, got:\n%s", want, src)
		}
	}
	if !strings.Contains(src, "session_start") || !strings.Contains(src, "sendMessage") {
		t.Fatal("extension should inject on session_start via sendMessage")
	}
}

func TestPIApplyIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	paths := []string{"AGENTS.md", "TODO.md"}
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
	paths := []string{"AGENTS.md", "TODO.md"}

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
	line := (PI{}).PlanLine([]string{"AGENTS.md", "TODO.md"})
	if !strings.Contains(line, ".pi/extensions/gyroscope-context.ts") {
		t.Fatalf("plan line should name the extension, got %q", line)
	}
}

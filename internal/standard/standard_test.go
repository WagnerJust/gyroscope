package standard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WagnerJust/gyroscope/internal/config"
)

func TestWriteAllSpokesLandsFilesAndGitignoresLocal(t *testing.T) {
	dir := t.TempDir()
	files, err := Plan(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Write(dir, files, false); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"AGENTS.md", "CONTEXT.md", ".local/local.md", "docs/agents.md", "docs/adr/TEMPLATE.md", "docs/agents/README.md"} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if got := string(gi); got == "" || !contains(got, ".local/") {
		t.Errorf(".gitignore should list .local/, got %q", got)
	}
}

func TestWriteRefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("mine"), 0o644)
	files, _ := Plan(config.Default())
	if _, err := Write(dir, files, false); err == nil {
		t.Fatal("expected refuse-to-overwrite error")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md")); string(b) != "mine" {
		t.Fatal("existing file must be untouched")
	}
}

func TestPlanDropsDisabledSpoke(t *testing.T) {
	cfg := config.Default()
	cfg.Spokes.Local = false
	files, _ := Plan(cfg)
	for _, f := range files {
		if f.Dest == ".local/local.md" {
			t.Fatal("local spoke should be dropped")
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WagnerJust/gyroscope/internal/config"
)

func runAgentsSet(t *testing.T, repo string, args ...string) error {
	t.Helper()
	cmd := newAgentsCmd(os.Stdout)
	cmd.SetArgs(append([]string{"set"}, append(args, repo)...))
	return cmd.Execute()
}

func TestAgentsSetWritesState(t *testing.T) {
	repo := t.TempDir()
	if err := runAgentsSet(t, repo, "skipped"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(repo)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Spokes.Personas != config.PersonaSkipped {
		t.Fatalf("got %q, want skipped", cfg.Spokes.Personas)
	}
}

func TestAgentsSetIsIdempotentAndOverwrites(t *testing.T) {
	repo := t.TempDir()
	if err := runAgentsSet(t, repo, "on"); err != nil {
		t.Fatal(err)
	}
	if err := runAgentsSet(t, repo, "skipped"); err != nil {
		t.Fatal(err) // must overwrite an existing gyroscope.json, not refuse
	}
	cfg, _ := config.Load(repo)
	if cfg.Spokes.Personas != config.PersonaSkipped {
		t.Fatalf("second set should win, got %q", cfg.Spokes.Personas)
	}
}

func TestAgentsSetPreservesOtherSpokes(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "gyroscope.json"), []byte(`{"spokes":{"local":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runAgentsSet(t, repo, "on"); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load(repo)
	if cfg.Spokes.Local {
		t.Fatal("local spoke should stay off after agents set")
	}
	if cfg.Spokes.Personas != config.PersonaOn {
		t.Fatalf("personas should be on, got %q", cfg.Spokes.Personas)
	}
}

func TestAgentsSetRejectsUnknownState(t *testing.T) {
	repo := t.TempDir()
	if err := runAgentsSet(t, repo, "banana"); err == nil {
		t.Fatal("expected error for unknown state")
	}
}

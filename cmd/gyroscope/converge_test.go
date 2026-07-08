package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/standard"
)

// stateOf returns the classified state for dest, failing if it is not present.
func stateOf(t *testing.T, items []convergeItem, dest string) convergeState {
	t.Helper()
	for _, it := range items {
		if it.Dest == dest {
			return it.State
		}
	}
	t.Fatalf("dest %q not classified; items=%v", dest, items)
	return stateNew
}

func TestClassifyFreshRepoAllNew(t *testing.T) {
	dir := t.TempDir()
	items, err := classifyAll(dir, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.State != stateNew {
			t.Errorf("%s: want NEW in a fresh repo, got %s", it.Dest, it.State)
		}
	}
	if len(conflicts(items)) != 0 {
		t.Fatalf("fresh repo should have no conflicts, got %v", conflicts(items))
	}
}

func TestClassifyIdenticalFilesAreOK(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir, "--apply"}, &out, &errb); err != nil {
		t.Fatalf("init --apply: %v (%s)", err, errb.String())
	}
	items, err := classifyAll(dir, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	// Every file init just wrote is byte-identical to what it would write again.
	for _, it := range items {
		if it.State != stateOK {
			t.Errorf("%s: want OK after a fresh init, got %s", it.Dest, it.State)
		}
	}
}

func TestClassifyConflictOnDivergentPlainFile(t *testing.T) {
	dir := t.TempDir()
	// A pointer file with foreign content and no managed region → CONFLICT.
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("my own notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := classifyAll(dir, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if got := stateOf(t, items, "GEMINI.md"); got != stateConflict {
		t.Fatalf("divergent pointer should be CONFLICT, got %s", got)
	}
}

func TestInitDryRunPrintsPerFileState(t *testing.T) {
	dir := t.TempDir()
	// One CONFLICT (foreign pointer), the rest NEW.
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := run([]string{"init", dir}, &out, &errb); err != nil {
		t.Fatalf("init dry-run: %v (%s)", err, errb.String())
	}
	s := out.String()
	if !strings.Contains(s, "NEW") || !strings.Contains(s, "AGENTS.md") {
		t.Fatalf("dry-run should classify AGENTS.md NEW, got:\n%s", s)
	}
	if !strings.Contains(s, "CONFLICT") || !strings.Contains(s, "GEMINI.md") {
		t.Fatalf("dry-run should classify the foreign GEMINI.md CONFLICT, got:\n%s", s)
	}
}

func TestClassifyMergeOnHubMissingManagedContent(t *testing.T) {
	dir := t.TempDir()
	// A hand-authored hub whose managed region has drifted (here: absent). The
	// classifier should offer to inject the managed region — MERGE, not CONFLICT.
	got := "# AGENTS.md\n\nSome user preamble gyroscope must not touch.\n\n" +
		standard.ManagedOpen + "\nstale — will be replaced\n" + standard.ManagedClose + "\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(got), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := classifyAll(dir, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if s := stateOf(t, items, "AGENTS.md"); s != stateMerge {
		t.Fatalf("hub with a stale managed region should be MERGE, got %s", s)
	}
}

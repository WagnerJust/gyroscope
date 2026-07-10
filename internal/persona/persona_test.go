package persona

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// writeFile is a test helper that writes content to repoDir/rel, creating dirs.
func writeFile(t *testing.T, repoDir, rel, content string) {
	t.Helper()
	dest := filepath.Join(repoDir, rel)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const validPersona = `---
name: analytics-dev
description: Use for the analytics layer.
tools: Bash, Read, Edit
model: sonnet
---

You are the analytics-dev persona. Do analytics things.
`

const anotherValidPersona = `---
name: verifier
description: Verify things.
---

You are the verifier.
`

func TestPlan_ValidPersonasOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "docs/agents/analytics-dev.md", validPersona)
	writeFile(t, dir, "docs/agents/verifier.md", anotherValidPersona)
	// README.md must be excluded even though it lives in docs/agents/.
	writeFile(t, dir, "docs/agents/README.md", "# Repo agent personas\n\nno frontmatter here\n")
	// A file lacking `name:` frontmatter is not a valid persona → excluded.
	writeFile(t, dir, "docs/agents/notes.md", "---\ndescription: no name here\n---\n\nbody\n")

	mirrors, err := Plan(dir)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	var dests []string
	for _, m := range mirrors {
		dests = append(dests, m.Dest)
	}
	sort.Strings(dests)

	want := []string{
		filepath.Join(".claude", "agents", "analytics-dev.md"),
		filepath.Join(".claude", "agents", "verifier.md"),
	}
	if len(dests) != len(want) {
		t.Fatalf("Plan dests = %v, want %v", dests, want)
	}
	for i := range want {
		if dests[i] != want[i] {
			t.Errorf("Plan dest[%d] = %q, want %q", i, dests[i], want[i])
		}
	}
}

func TestPlan_DestUsesFrontmatterName(t *testing.T) {
	dir := t.TempDir()
	// The source FILE is named differently from the frontmatter `name:` — the dest
	// filename must come from `name:`, not the source filename.
	writeFile(t, dir, "docs/agents/source-file.md", validPersona)

	mirrors, err := Plan(dir)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(mirrors) != 1 {
		t.Fatalf("Plan returned %d mirrors, want 1", len(mirrors))
	}
	wantDest := filepath.Join(".claude", "agents", "analytics-dev.md")
	if mirrors[0].Dest != wantDest {
		t.Errorf("dest = %q, want %q (from frontmatter name)", mirrors[0].Dest, wantDest)
	}
	if mirrors[0].Src != filepath.Join("docs", "agents", "source-file.md") {
		t.Errorf("src = %q, want docs/agents/source-file.md", mirrors[0].Src)
	}
}

func TestMirror_WritesByteEqualCopies(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "docs/agents/analytics-dev.md", validPersona)
	writeFile(t, dir, "docs/agents/verifier.md", anotherValidPersona)
	writeFile(t, dir, "docs/agents/README.md", "# readme, not a persona\n")

	written, err := Copy(dir)
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if len(written) != 2 {
		t.Fatalf("Copy wrote %d files, want 2: %v", len(written), written)
	}

	// Dest bytes must equal source bytes, byte-for-byte.
	for _, rel := range []string{"analytics-dev.md", "verifier.md"} {
		dst := filepath.Join(dir, ".claude", "agents", rel)
		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("read mirror %s: %v", rel, err)
		}
		// The source may have a different filename for analytics-dev? Not here — the
		// frontmatter name matches the source filename in this test.
		var srcName string
		if rel == "analytics-dev.md" {
			srcName = "analytics-dev.md"
		} else {
			srcName = "verifier.md"
		}
		want, err := os.ReadFile(filepath.Join(dir, "docs", "agents", srcName))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("mirror %s not byte-equal to source", rel)
		}
	}

	// README.md was NOT mirrored.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "README.md")); !os.IsNotExist(err) {
		t.Errorf(".claude/agents/README.md should not exist (README is not a persona)")
	}
}

func TestMirror_OverwritesOnDrift(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "docs/agents/analytics-dev.md", validPersona)

	if _, err := Copy(dir); err != nil {
		t.Fatalf("first Copy: %v", err)
	}
	// Simulate drift: someone edited the registered mirror.
	dst := filepath.Join(dir, ".claude", "agents", "analytics-dev.md")
	if err := os.WriteFile(dst, []byte("stale edited content"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Re-mirror must overwrite the gyroscope-owned mirror (no O_EXCL refusal).
	if _, err := Copy(dir); err != nil {
		t.Fatalf("second Copy: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != validPersona {
		t.Errorf("re-mirror did not overwrite drifted file; got %q", got)
	}
}

func TestMirror_NoPersonasIsNoOp(t *testing.T) {
	dir := t.TempDir()
	// Only a README, no personas.
	writeFile(t, dir, "docs/agents/README.md", "# readme\n")

	written, err := Copy(dir)
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if len(written) != 0 {
		t.Errorf("Copy wrote %v, want nothing", written)
	}
	// .claude/agents/ should not be created when there is nothing to mirror.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents")); !os.IsNotExist(err) {
		t.Errorf(".claude/agents/ should not be created for a no-op mirror")
	}
}

func TestMirror_MissingDocsAgentsIsNoOp(t *testing.T) {
	dir := t.TempDir()
	written, err := Copy(dir)
	if err != nil {
		t.Fatalf("Copy on missing docs/agents/: %v", err)
	}
	if len(written) != 0 {
		t.Errorf("Copy wrote %v, want nothing", written)
	}
}

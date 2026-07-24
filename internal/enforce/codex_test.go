package enforce

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readCodex(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, codexConfigPath))
	if err != nil {
		t.Fatalf("read codex config: %v", err)
	}
	return string(b)
}

func TestSetCodexAttribution_CreatesFileWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	changed, err := SetCodexCommitAttributionEmpty(dir)
	if err != nil || !changed {
		t.Fatalf("expected changed, got changed=%v err=%v", changed, err)
	}
	if got := readCodex(t, dir); !strings.Contains(got, `commit_attribution = ""`) {
		t.Errorf("created config missing the key:\n%s", got)
	}
	val, present, err := CodexCommitAttribution(dir)
	if err != nil || !present || val != "" {
		t.Errorf("read-back: val=%q present=%v err=%v", val, present, err)
	}
}

func TestSetCodexAttribution_ReplacesExistingValuePreservingRest(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := "# my codex config\n" +
		`commit_attribution = "Codex <noreply@openai.com>"` + "\n" +
		"\n[features]\ncodex_git_commit = true\n"
	if err := os.WriteFile(filepath.Join(dir, codexConfigPath), []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := SetCodexCommitAttributionEmpty(dir)
	if err != nil || !changed {
		t.Fatalf("expected changed, got changed=%v err=%v", changed, err)
	}
	got := readCodex(t, dir)
	if !strings.Contains(got, `commit_attribution = ""`) {
		t.Errorf("value not replaced:\n%s", got)
	}
	// Other content preserved.
	for _, keep := range []string{"# my codex config", "[features]", "codex_git_commit = true"} {
		if !strings.Contains(got, keep) {
			t.Errorf("edit dropped preserved content %q:\n%s", keep, got)
		}
	}
	// Only one attribution line.
	if n := strings.Count(got, "commit_attribution"); n != 1 {
		t.Errorf("expected exactly one commit_attribution line, got %d:\n%s", n, got)
	}
}

func TestSetCodexAttribution_PrependsWhenKeyAbsent(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := "[features]\ncodex_git_commit = true\n"
	if err := os.WriteFile(filepath.Join(dir, codexConfigPath), []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SetCodexCommitAttributionEmpty(dir); err != nil {
		t.Fatal(err)
	}
	got := readCodex(t, dir)
	// The key must land above the first table so it is genuinely top-level.
	if !strings.HasPrefix(got, `commit_attribution = ""`) {
		t.Errorf("key should be prepended above the first table:\n%s", got)
	}
	if !strings.Contains(got, "[features]") {
		t.Errorf("existing table dropped:\n%s", got)
	}
}

func TestSetCodexAttribution_AlreadyEmptyIsNoOp(t *testing.T) {
	dir := t.TempDir()
	if _, err := SetCodexCommitAttributionEmpty(dir); err != nil {
		t.Fatal(err)
	}
	changed, err := SetCodexCommitAttributionEmpty(dir)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("second call should be a no-op (already empty)")
	}
}

// A commit_attribution key INSIDE a table is a different key (table.commit_attribution)
// and must not be read as the top-level one.
func TestCodexAttribution_IgnoresKeyInsideTable(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := "[some_table]\n" + `commit_attribution = "not top level"` + "\n"
	if err := os.WriteFile(filepath.Join(dir, codexConfigPath), []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	_, present, err := CodexCommitAttribution(dir)
	if err != nil {
		t.Fatal(err)
	}
	if present {
		t.Error("a key inside a table must not count as the top-level commit_attribution")
	}
}

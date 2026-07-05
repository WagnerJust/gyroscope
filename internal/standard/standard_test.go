package standard

import (
	"os"
	"path/filepath"
	"strings"
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
	if got := string(gi); got == "" || !strings.Contains(got, ".local/") {
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

func agentsContent(t *testing.T, files []File) string {
	t.Helper()
	for _, f := range files {
		if f.Dest == "AGENTS.md" {
			return string(f.Content)
		}
	}
	t.Fatal("AGENTS.md not present in plan")
	return ""
}

func TestPlanCustomSpokeEmitsFileAndRoute(t *testing.T) {
	cfg := config.Default()
	cfg.Custom = []config.CustomSpoke{{Name: "Security", Dest: "docs/security.md"}}
	files, err := Plan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var stub *File
	for i := range files {
		if files[i].Dest == "docs/security.md" {
			stub = &files[i]
		}
	}
	if stub == nil {
		t.Fatal("custom spoke should emit a file at docs/security.md")
	}
	if !strings.Contains(string(stub.Content), "# Security") {
		t.Fatalf("custom stub should be a scaffold titled with the name, got %q", stub.Content)
	}
	agents := agentsContent(t, files)
	if !strings.Contains(agents, "docs/security.md") || !strings.Contains(agents, "**Security**") {
		t.Fatalf("hub should route to the custom spoke, got:\n%s", agents)
	}
	if strings.Contains(agents, "gyroscope:custom-routes") {
		t.Fatalf("hub must not still contain the raw marker, got:\n%s", agents)
	}
}

func TestPlanNoCustomSpokesRemovesMarker(t *testing.T) {
	files, err := Plan(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	agents := agentsContent(t, files)
	if strings.Contains(agents, "gyroscope:custom-routes") {
		t.Fatalf("hub must not contain the marker when there are no custom spokes, got:\n%s", agents)
	}
	// Sanity: still the built-in hub with its Routes and Pointer files sections.
	if !strings.Contains(agents, "## Routes") || !strings.Contains(agents, "## Pointer files") {
		t.Fatalf("hub should remain the built-in hub, got:\n%s", agents)
	}
	if strings.Contains(agents, "\n\n\n") {
		t.Fatalf("marker removal should not leave a double blank line, got:\n%s", agents)
	}
}

func TestPlanSkipsCustomSpokeMissingNameOrDest(t *testing.T) {
	cfg := config.Default()
	cfg.Custom = []config.CustomSpoke{{Name: "", Dest: "docs/x.md"}, {Name: "Y", Dest: ""}}
	files, err := Plan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if f.Dest == "docs/x.md" {
			t.Fatal("custom spoke with empty name should be skipped")
		}
	}
	agents := agentsContent(t, files)
	if strings.Contains(agents, "gyroscope:custom-routes") {
		t.Fatalf("marker should be gone even when all custom entries are skipped, got:\n%s", agents)
	}
}

func TestWriteForceOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := Plan(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Write(dir, files, true); err != nil {
		t.Fatalf("force write should succeed over existing file: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) == "mine" {
		t.Fatal("force should have overwritten the existing file")
	}
	var want []byte
	for _, f := range files {
		if f.Dest == "AGENTS.md" {
			want = f.Content
		}
	}
	if string(got) != string(want) {
		t.Fatalf("force-written AGENTS.md should match template content")
	}
}

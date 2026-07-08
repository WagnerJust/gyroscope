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

func TestPlanIncludesL2ProcessArtifacts(t *testing.T) {
	files, err := Plan(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	content := map[string][]byte{}
	for _, f := range files {
		content[f.Dest] = f.Content
	}
	for _, want := range []string{".github/pull_request_template.md", ".gitmessage"} {
		if _, ok := content[want]; !ok {
			t.Errorf("default plan should include %s", want)
		}
	}
	// plumbline's l2.pr-template scores Found only with ≥3 markdown checkboxes.
	if n := strings.Count(string(content[".github/pull_request_template.md"]), "- [ ]"); n < 3 {
		t.Errorf("PR template needs ≥3 `- [ ]` checkboxes for plumbline Found, got %d", n)
	}
}

func TestPlanRoutesContributingFromHub(t *testing.T) {
	files, err := Plan(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	var body []byte
	for _, f := range files {
		if f.Dest == "CONTRIBUTING.md" {
			body = f.Content
		}
	}
	if body == nil {
		t.Fatal("default plan should include CONTRIBUTING.md")
	}
	// plumbline's l2.contributor-guide needs a heading + ≥20 non-blank lines.
	var nonBlank int
	for _, ln := range strings.Split(string(body), "\n") {
		if strings.TrimSpace(ln) != "" {
			nonBlank++
		}
	}
	if nonBlank < 20 {
		t.Errorf("CONTRIBUTING needs ≥20 non-blank lines for plumbline Found, got %d", nonBlank)
	}
	if !strings.Contains(string(body), "#") {
		t.Error("CONTRIBUTING needs a heading for plumbline Found")
	}
	// No-drift guarantee: CONTRIBUTING defers conventions to docs/agents.md
	// rather than copying them.
	if !strings.Contains(string(body), "docs/agents.md") {
		t.Error("CONTRIBUTING should defer to docs/agents.md, not duplicate conventions")
	}
	agents := agentsContent(t, files)
	if !strings.Contains(agents, "CONTRIBUTING.md") {
		t.Fatalf("hub should route to CONTRIBUTING.md, got:\n%s", agents)
	}
}

func TestPlanIncludesStateFiles(t *testing.T) {
	files, err := Plan(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range files {
		got[f.Dest] = true
	}
	for _, want := range []string{"TODO.md", ".local/todo.md"} {
		if !got[want] {
			t.Errorf("default plan should include the state file %s", want)
		}
	}
	agents := agentsContent(t, files)
	if !strings.Contains(agents, "TODO.md") {
		t.Fatalf("hub should route to TODO.md, got:\n%s", agents)
	}
}

func TestHubCarriesPersonasDirectiveWhenEnabled(t *testing.T) {
	cfg := config.Default() // personas = unknown → enabled
	files, err := Plan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	hub := agentsContent(t, files)
	if !strings.Contains(hub, "spokes.personas") {
		t.Fatal("hub should carry the personas directive when the spoke is enabled")
	}
	if strings.Contains(hub, "<!-- gyroscope:personas-directive -->") {
		t.Fatal("the personas marker should be replaced, not left in the hub")
	}
}

func TestHubOmitsPersonasDirectiveWhenOff(t *testing.T) {
	cfg := config.Default()
	cfg.Spokes.Personas = config.PersonaOff
	files, err := Plan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	hub := agentsContent(t, files)
	if strings.Contains(hub, "spokes.personas") {
		t.Fatal("hub should not carry the personas directive when the spoke is off")
	}
	if strings.Contains(hub, "<!-- gyroscope:personas-directive -->") {
		t.Fatal("the personas marker should be removed when off")
	}
}

func TestHubOmitsDisabledSpokeRoutes(t *testing.T) {
	cfg := config.Default()
	cfg.Spokes.Context = false
	cfg.Spokes.ADR = false
	cfg.Spokes.Personas = config.PersonaOff
	files, err := Plan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	agents := agentsContent(t, files)
	// Disabled spokes must not be routed — following the hub never hits a file
	// that wasn't written.
	for _, gone := range []string{"CONTEXT.md", "docs/adr/", "docs/agents/"} {
		if strings.Contains(agents, gone) {
			t.Errorf("disabled spoke route %q must be pruned from the hub, got:\n%s", gone, agents)
		}
	}
	// Enabled spokes stay routed.
	for _, want := range []string{"docs/agents.md", "TODO.md", "CONTRIBUTING.md"} {
		if !strings.Contains(agents, want) {
			t.Errorf("enabled spoke route %q should still be present", want)
		}
	}
	// The now-unnecessary "spokes are optional" hedge is gone.
	if strings.Contains(agents, "Spokes are optional") {
		t.Errorf("hedge should be removed now that routes are pruned, got:\n%s", agents)
	}
}

func TestPlanDropsDisabledState(t *testing.T) {
	cfg := config.Default()
	cfg.Spokes.State = false
	files, _ := Plan(cfg)
	for _, f := range files {
		if f.Dest == "TODO.md" || f.Dest == ".local/todo.md" {
			t.Fatalf("disabled state spoke should be dropped: %s", f.Dest)
		}
	}
}

func TestPlanDropsDisabledContributing(t *testing.T) {
	cfg := config.Default()
	cfg.Spokes.Contributing = false
	files, _ := Plan(cfg)
	for _, f := range files {
		if f.Dest == "CONTRIBUTING.md" {
			t.Fatal("disabled contributing spoke should be dropped")
		}
	}
}

func TestPlanDropsDisabledProcessArtifacts(t *testing.T) {
	cfg := config.Default()
	cfg.Spokes.PRTemplate = false
	cfg.Spokes.CommitConvention = false
	files, _ := Plan(cfg)
	for _, f := range files {
		if f.Dest == ".github/pull_request_template.md" || f.Dest == ".gitmessage" {
			t.Fatalf("disabled process artifact should be dropped: %s", f.Dest)
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
	if strings.Contains(agents, "gyroscope:routes") {
		t.Fatalf("hub must not still contain the raw marker, got:\n%s", agents)
	}
}

func TestPlanNoCustomSpokesRemovesMarker(t *testing.T) {
	files, err := Plan(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	agents := agentsContent(t, files)
	if strings.Contains(agents, "gyroscope:routes") {
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
	if strings.Contains(agents, "gyroscope:routes") {
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

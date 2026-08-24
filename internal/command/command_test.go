package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCmd(t *testing.T, repo, name, content string) {
	t.Helper()
	dir := filepath.Join(repo, "docs", "commands")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A full-frontmatter command carrying every Claude key + portable body syntax.
const fullCmd = "---\n" +
	"description: Fix a GitHub issue\n" +
	"argument-hint: [issue-number]\n" +
	"allowed-tools: Bash(git *), Read\n" +
	"model: claude-sonnet-4-5\n" +
	"---\n" +
	"Fix issue $ARGUMENTS. Run !`git status` and read @README.md first.\n"

func TestParse_FrontmatterAndBody(t *testing.T) {
	c := parse("fix", []byte(fullCmd))
	if c.front("description") != "Fix a GitHub issue" {
		t.Errorf("description = %q", c.front("description"))
	}
	if c.front("argument-hint") != "[issue-number]" {
		t.Errorf("argument-hint = %q", c.front("argument-hint"))
	}
	if !strings.Contains(c.Body, "$ARGUMENTS") || strings.Contains(c.Body, "description:") {
		t.Errorf("body must start after frontmatter: %q", c.Body)
	}
}

func TestParse_NoFrontmatterIsWholeBody(t *testing.T) {
	c := parse("bare", []byte("Just a prompt, no frontmatter.\n"))
	if len(c.Front) != 0 || !strings.Contains(c.Body, "Just a prompt") {
		t.Errorf("bare command mishandled: front=%v body=%q", c.Front, c.Body)
	}
}

func TestRenderClaude_Verbatim(t *testing.T) {
	c := parse("fix", []byte(fullCmd))
	if string(Claude.render(c)) != fullCmd {
		t.Error("Claude render must be a verbatim copy of the canonical source")
	}
}

func TestRenderOpenCode_KeepsDescriptionDropsClaudeOnly(t *testing.T) {
	c := parse("fix", []byte(fullCmd))
	got := string(OpenCode.render(c))
	if !strings.Contains(got, "description: Fix a GitHub issue") {
		t.Errorf("opencode must keep description:\n%s", got)
	}
	for _, gone := range []string{"argument-hint:", "allowed-tools:", "model:"} {
		if strings.Contains(got, gone) {
			t.Errorf("opencode must drop %q:\n%s", gone, got)
		}
	}
	// Body (with portable $ARGUMENTS / ! / @) kept verbatim.
	if !strings.Contains(got, "Fix issue $ARGUMENTS. Run !`git status` and read @README.md first.") {
		t.Errorf("opencode must keep the body verbatim:\n%s", got)
	}
}

func TestRenderPI_KeepsDescriptionAndArgHint(t *testing.T) {
	c := parse("fix", []byte(fullCmd))
	got := string(PI.render(c))
	if !strings.Contains(got, "description: Fix a GitHub issue") || !strings.Contains(got, "argument-hint: [issue-number]") {
		t.Errorf("pi must keep description + argument-hint:\n%s", got)
	}
	for _, gone := range []string{"allowed-tools:", "model:"} {
		if strings.Contains(got, gone) {
			t.Errorf("pi must drop %q:\n%s", gone, got)
		}
	}
}

func TestRenderKeep_NoKeptKeysEmitsBodyOnly(t *testing.T) {
	// A command whose only frontmatter is a Claude-only key → OpenCode emits body only.
	c := parse("x", []byte("---\nmodel: claude-sonnet-4-5\n---\nBody only.\n"))
	got := string(OpenCode.render(c))
	if strings.HasPrefix(got, "---") {
		t.Errorf("no kept keys should emit no frontmatter block:\n%s", got)
	}
	if !strings.Contains(got, "Body only.") {
		t.Errorf("body lost:\n%s", got)
	}
}

func TestPlanFor_PerTargetDestAndOrder(t *testing.T) {
	dir := t.TempDir()
	writeCmd(t, dir, "deploy.md", fullCmd)
	writeCmd(t, dir, "test.md", "Run the tests.\n")

	cases := []struct {
		target   Target
		wantDest string
	}{
		{Claude, filepath.Join(".claude", "commands", "deploy.md")},
		{OpenCode, filepath.Join(".opencode", "commands", "deploy.md")},
		{PI, filepath.Join(".pi", "prompts", "deploy.md")},
	}
	for _, tc := range cases {
		mirrors, err := PlanFor(dir, tc.target)
		if err != nil {
			t.Fatalf("%s: %v", tc.target.ID, err)
		}
		if len(mirrors) != 2 {
			t.Fatalf("%s: got %d mirrors, want 2", tc.target.ID, len(mirrors))
		}
		// os.ReadDir is sorted: deploy.md before test.md.
		if mirrors[0].Dest != tc.wantDest {
			t.Errorf("%s: dest[0] = %q, want %q", tc.target.ID, mirrors[0].Dest, tc.wantDest)
		}
	}
}

func TestPlanFor_MissingDirIsEmpty(t *testing.T) {
	m, err := PlanFor(t.TempDir(), Claude)
	if err != nil || len(m) != 0 {
		t.Fatalf("missing docs/commands: m=%v err=%v", m, err)
	}
}

package persona

import (
	"path/filepath"
	"strings"
	"testing"
)

// personaWithQuotes exercises the tricky cases: a description carrying embedded
// double quotes (real personas do) and a body carrying a code fence.
const personaWithQuotes = "---\n" +
	"name: backend-architect\n" +
	`description: Designs things. Use for "how should this be structured", or review.` + "\n" +
	"tools: Read, Grep, Bash\n" +
	"model: sonnet\n" +
	"---\n" +
	"\n" +
	"# Backend Architect\n\nYou design things. Prefer `WriteGuarded`.\n"

func TestParse_ExtractsNameDescriptionBody(t *testing.T) {
	p, ok := parse([]byte(personaWithQuotes))
	if !ok {
		t.Fatal("valid persona should parse")
	}
	if p.Name != "backend-architect" {
		t.Errorf("name = %q", p.Name)
	}
	if !strings.Contains(p.Description, `"how should this be structured"`) {
		t.Errorf("description lost embedded quotes: %q", p.Description)
	}
	if !strings.Contains(p.Body, "You design things") {
		t.Errorf("body = %q", p.Body)
	}
	if strings.Contains(p.Body, "name:") {
		t.Errorf("body must start AFTER frontmatter, got: %q", p.Body)
	}
}

func TestPlanFor_PerTargetDestAndEncoding(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "docs/agents/backend-architect.md", personaWithQuotes)

	cases := []struct {
		target   Target
		wantDest string
	}{
		{Claude, filepath.Join(".claude", "agents", "backend-architect.md")},
		{OpenCode, filepath.Join(".opencode", "agents", "backend-architect.md")},
		{Codex, filepath.Join(".codex", "agents", "backend-architect.toml")},
	}
	for _, c := range cases {
		mirrors, err := PlanFor(dir, c.target)
		if err != nil {
			t.Fatalf("%s: PlanFor: %v", c.target.ID, err)
		}
		if len(mirrors) != 1 {
			t.Fatalf("%s: got %d mirrors, want 1", c.target.ID, len(mirrors))
		}
		if mirrors[0].Dest != c.wantDest {
			t.Errorf("%s: dest = %q, want %q", c.target.ID, mirrors[0].Dest, c.wantDest)
		}
	}
}

func TestRenderClaude_IsByteCopy(t *testing.T) {
	p, _ := parse([]byte(personaWithQuotes))
	if got := string(Claude.render(p)); got != personaWithQuotes {
		t.Errorf("Claude render must be a byte-copy of the source, got:\n%s", got)
	}
}

func TestRenderOpenCode_FrontmatterAndBody(t *testing.T) {
	p, _ := parse([]byte(personaWithQuotes))
	got := string(OpenCode.render(p))
	if !strings.HasPrefix(got, "---\n") {
		t.Fatalf("OpenCode output must start with frontmatter, got:\n%s", got)
	}
	if !strings.Contains(got, "mode: subagent") {
		t.Errorf("OpenCode agent must declare mode: subagent:\n%s", got)
	}
	if !strings.Contains(got, `description: Designs things. Use for "how should this be structured", or review.`) {
		t.Errorf("OpenCode must carry the description verbatim:\n%s", got)
	}
	// Claude-only keys must be dropped.
	for _, gone := range []string{"name:", "tools:", "model:"} {
		if strings.Contains(frontmatterOf(got), gone) {
			t.Errorf("OpenCode frontmatter should drop %q:\n%s", gone, got)
		}
	}
	// Body preserved verbatim, including the code fence backticks.
	if !strings.Contains(got, "You design things. Prefer `WriteGuarded`.") {
		t.Errorf("OpenCode must keep the body verbatim:\n%s", got)
	}
}

func TestRenderCodex_IsValidTOMLWithEscaping(t *testing.T) {
	p, _ := parse([]byte(personaWithQuotes))
	got := string(Codex.render(p))
	if !strings.Contains(got, `name = "backend-architect"`) {
		t.Errorf("codex name field missing:\n%s", got)
	}
	// The embedded double quotes in the description must be escaped for a TOML
	// single-line basic string.
	if !strings.Contains(got, `\"how should this be structured\"`) {
		t.Errorf("codex description quotes not escaped:\n%s", got)
	}
	// The body becomes a multi-line basic string field, not the file body.
	if !strings.Contains(got, `developer_instructions = """`) {
		t.Errorf("codex must put the prompt in developer_instructions:\n%s", got)
	}
	if !strings.Contains(got, "You design things") {
		t.Errorf("codex must carry the prompt body:\n%s", got)
	}
}

func TestTomlMultiline_NeutralizesTripleQuoteAndBackslash(t *testing.T) {
	// A body that contains a literal """ and a backslash must not break the string.
	out := tomlMultiline(`ends with a quote block """ and a path C:\x`)
	if strings.Contains(out, `"""`+" and") { // the inner """ must have been escaped
		t.Errorf("inner triple-quote not neutralized: %s", out)
	}
	if !strings.HasSuffix(out, `"""`) {
		t.Errorf("must close with a triple quote: %s", out)
	}
	if !strings.Contains(out, `C:\\x`) {
		t.Errorf("backslash not escaped: %s", out)
	}
}

// frontmatterOf returns the text between the first two `---` lines of a mirror.
func frontmatterOf(s string) string {
	rest, ok := strings.CutPrefix(s, "---\n")
	if !ok {
		return ""
	}
	if i := strings.Index(rest, "\n---\n"); i >= 0 {
		return rest[:i]
	}
	return rest
}

package enforce

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeSatisfiesAdapter(t *testing.T) {
	var _ Adapter = Claude{}
}

func TestClaudeApplyThenVerify(t *testing.T) {
	dir := t.TempDir()
	c := Claude{}
	p := Paths{Rules: []string{"AGENTS.md", "docs/agents.md"}, State: []string{"TODO.md", ".local/todo.md"}}

	changed, err := c.Apply(dir, p)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("first Apply should install the hook")
	}
	ok, err := c.Verify(dir, p)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Verify should see the hook Apply just installed")
	}
	changed, err = c.Apply(dir, p)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second Apply should be idempotent")
	}
}

// The Claude hook carries only dynamic state: the rules reach Claude Code through
// CLAUDE.md's @imports, so they must not be catted by the hook.
func TestClaudeHookCarriesStateOnly(t *testing.T) {
	dir := t.TempDir()
	p := Paths{Rules: []string{"AGENTS.md", "docs/agents.md"}, State: []string{"TODO.md"}}
	if _, err := (Claude{}).Apply(dir, p); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if ok, _ := (Claude{}).HasSessionStart(dir, SessionStartCommand(p.State)); !ok {
		t.Fatalf("settings should carry the state-only command, got: %s", b)
	}
	if strings.Contains(string(b), "docs/agents.md") || strings.Contains(string(b), "cat AGENTS.md") {
		t.Fatalf("hook must not inject rules files, got: %s", b)
	}
}

// With no state spokes enabled there is nothing for the hook to carry: Apply
// installs nothing (and removes a stale gyroscope hook), Verify is true iff no
// gyroscope-shaped hook remains.
func TestClaudeEmptyStateInstallsNoHook(t *testing.T) {
	dir := t.TempDir()
	c := Claude{}
	empty := Paths{Rules: []string{"AGENTS.md"}}

	changed, err := c.Apply(dir, empty)
	if err != nil || changed {
		t.Fatalf("Apply with no state on a bare repo: changed=%v err=%v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Fatal("no hook to install: settings.json must not be created")
	}
	if ok, err := c.Verify(dir, empty); err != nil || !ok {
		t.Fatalf("Verify with no state and no hook: ok=%v err=%v", ok, err)
	}

	// A hook from a previous config (state on) is now stale.
	if _, err := c.Apply(dir, Paths{State: []string{"TODO.md"}}); err != nil {
		t.Fatal(err)
	}
	if ok, _ := c.Verify(dir, empty); ok {
		t.Fatal("Verify should be false while a gyroscope hook is present and state is empty")
	}
	changed, err = c.Apply(dir, empty)
	if err != nil || !changed {
		t.Fatalf("Apply should remove the stale hook: changed=%v err=%v", changed, err)
	}
	if ok, _ := c.Verify(dir, empty); !ok {
		t.Fatal("Verify should be true once the stale hook is removed")
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if strings.Contains(string(b), "gyroscope-hook") {
		t.Fatalf("stale hook still present: %s", b)
	}
}

func TestClaudePlanLineSummarizesTheHook(t *testing.T) {
	line := (Claude{}).PlanLine(Paths{
		Rules: []string{"AGENTS.md"},
		State: []string{"gyroscope.json", "TODO.md", ".local/todo.md", ".local/local.md"},
	})
	want := "merge: .claude/settings.json — SessionStart hook injects state within 9000 bytes: gyroscope.json TODO.md .local/todo.md .local/local.md"
	if line != want {
		t.Fatalf("plan line:\n got %q\nwant %q", line, want)
	}
}

func TestSessionStartCommandShape(t *testing.T) {
	got := SessionStartCommand([]string{"gyroscope.json", "TODO.md", ".local/todo.md", ".local/local.md"})
	want := `b=9000; printf '%s\n' '## Session state (gyroscope): working files for resuming, injected within a 9000-byte budget. Repo rules live in AGENTS.md and the spokes it routes to, not here.'; for f in gyroscope.json TODO.md .local/todo.md .local/local.md; do [ -r "$f" ] || continue; s=$(wc -c <"$f" | tr -d ' '); if [ "$s" -le "$b" ]; then printf '\n### %s (%s bytes)\n' "$f" "$s"; cat "$f"; b=$((b-s)); else printf '\n### %s: %s bytes, over the remaining session budget; not injected. Read it when resuming.\n' "$f" "$s"; fi; done; true # gyroscope-hook`
	if got != want {
		t.Fatalf("command:\n got %s\nwant %s", got, want)
	}
	if strings.Contains(SessionStartCommand([]string{"TODO.md"}), "local.md") {
		t.Fatal("an omitted path must not appear in the command")
	}
}

// The command must survive the settings.json encoder byte-for-byte: no HTML
// escaping of <, >, or & (it contains `wc -c <"$f"`).
func TestSessionStartCommandSurvivesSettingsEncoder(t *testing.T) {
	dir := t.TempDir()
	cmd := SessionStartCommand([]string{"TODO.md"})
	if _, err := (Claude{}).Install(dir, cmd); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if strings.Contains(string(b), `\u003c`) || strings.Contains(string(b), `\u003e`) || strings.Contains(string(b), `\u0026`) {
		t.Fatalf("command was HTML-escaped: %s", b)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if ok, err := (Claude{}).HasSessionStart(dir, cmd); err != nil || !ok {
		t.Fatalf("round-trip: ok=%v err=%v", ok, err)
	}
	if !strings.Contains(string(b), `wc -c <\"$f\"`) {
		t.Fatalf("literal < not preserved: %s", b)
	}
}

// Install replaces gyroscope's own previous hooks rather than appending beside
// them: the legacy `cat AGENTS.md ...` shape and any older tagged command are
// removed, every foreign hook is untouched, and a second Install is a no-op.
func TestClaudeInstallReplacesLegacyHook(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	seed := `{"hooks":{"SessionStart":[` +
		`{"hooks":[{"type":"command","command":"cat AGENTS.md docs/agents.md TODO.md 2>/dev/null || true"}]},` +
		`{"matcher":"startup","hooks":[{"type":"command","command":"echo foreign"},{"type":"command","command":"b=1; true # gyroscope-hook"}]}` +
		`]}}`
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := SessionStartCommand([]string{"TODO.md"})
	changed, err := (Claude{}).Install(dir, cmd)
	if err != nil || !changed {
		t.Fatalf("install over legacy: changed=%v err=%v", changed, err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	list := m["hooks"].(map[string]any)["SessionStart"].([]any)
	var cmds []string
	for _, e := range list {
		inner := e.(map[string]any)["hooks"].([]any)
		if len(inner) == 0 {
			t.Fatalf("an emptied entry group must be dropped: %s", b)
		}
		for _, h := range inner {
			cmds = append(cmds, h.(map[string]any)["command"].(string))
		}
	}
	gyro, foreign := 0, 0
	for _, c := range cmds {
		switch {
		case c == cmd:
			gyro++
		case c == "echo foreign":
			foreign++
		default:
			t.Fatalf("unexpected leftover hook %q in %s", c, b)
		}
	}
	if gyro != 1 || foreign != 1 {
		t.Fatalf("want exactly one gyroscope hook and the foreign hook intact, got %v", cmds)
	}
	if !strings.Contains(string(b), `"matcher": "startup"`) {
		t.Fatalf("the foreign hook's group (and its matcher) must survive: %s", b)
	}

	changed, err = (Claude{}).Install(dir, cmd)
	if err != nil || changed {
		t.Fatalf("second install should be a no-op: changed=%v err=%v", changed, err)
	}
}

func TestClaudeInstallAddsHookThenIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	c := Claude{}
	cmd := SessionStartCommand([]string{"TODO.md", ".local/local.md"})

	changed, err := c.Install(dir, cmd)
	if err != nil || !changed {
		t.Fatalf("first install: changed=%v err=%v", changed, err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("settings not valid json: %v", err)
	}
	if ok, _ := c.HasSessionStart(dir, cmd); !ok {
		t.Fatalf("written settings should contain the hook command, got: %s", b)
	}

	changed, err = c.Install(dir, cmd)
	if err != nil || changed {
		t.Fatalf("second install should be a no-op: changed=%v err=%v", changed, err)
	}
}

func TestClaudeInstallPreservesExistingSettings(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude", "settings.json"),
		[]byte(`{"model":"opus","hooks":{"Stop":[{"hooks":[]}]}}`), 0o644)

	cmd := SessionStartCommand([]string{"TODO.md", ".local/local.md"})
	if _, err := (Claude{}).Install(dir, cmd); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	var m map[string]any
	json.Unmarshal(b, &m)
	if m["model"] != "opus" {
		t.Fatal("existing top-level setting was clobbered")
	}
	hooks := m["hooks"].(map[string]any)
	if _, ok := hooks["Stop"]; !ok {
		t.Fatal("existing Stop hook was clobbered")
	}
	if _, ok := hooks["SessionStart"]; !ok {
		t.Fatal("SessionStart hook not added")
	}
}

func TestClaudeInstallPreservesOtherToolsSessionStart(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	const other = "echo other-tool"
	seed := `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"echo other-tool"}]}]}}`
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := SessionStartCommand([]string{"TODO.md", ".local/local.md"})
	changed, err := (Claude{}).Install(dir, cmd)
	if err != nil || !changed {
		t.Fatalf("install onto a populated SessionStart: changed=%v err=%v", changed, err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	got := string(b)
	if !strings.Contains(got, other) {
		t.Fatalf("another tool's SessionStart hook was dropped, got: %s", got)
	}
	if ok, _ := (Claude{}).HasSessionStart(dir, cmd); !ok {
		t.Fatalf("gyroscope hook not appended, got: %s", got)
	}
}

func TestClaudeInstallRejectsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	path := filepath.Join(dir, ".claude", "settings.json")
	const corrupt = `{"model":"opus"` // truncated — not valid JSON
	if err := os.WriteFile(path, []byte(corrupt), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := SessionStartCommand([]string{"TODO.md", ".local/local.md"})
	if _, err := (Claude{}).Install(dir, cmd); err == nil {
		t.Fatal("expected an error on malformed settings.json")
	}
	if b, _ := os.ReadFile(path); string(b) != corrupt {
		t.Fatalf("malformed file must be left untouched, got: %s", b)
	}
}

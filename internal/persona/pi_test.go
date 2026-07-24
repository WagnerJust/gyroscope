package persona

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPIExtensionMirror_NoPersonasIsNotOk(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "docs/agents/README.md", "# readme\n")
	_, ok, err := PIExtensionMirror(dir)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("no personas → PIExtensionMirror should report ok=false (write nothing)")
	}
}

func TestPIExtensionMirror_RegistersToolPerPersona(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "docs/agents/backend-architect.md", personaWithQuotes)
	writeFile(t, dir, "docs/agents/verifier.md", anotherValidPersona)

	m, ok, err := PIExtensionMirror(dir)
	if err != nil || !ok {
		t.Fatalf("expected a mirror, ok=%v err=%v", ok, err)
	}
	if m.Dest != filepath.FromSlash(".pi/extensions/gyroscope-personas.ts") {
		t.Errorf("dest = %q", m.Dest)
	}
	src := string(m.Bytes)

	// The mechanism: registerTool + a headless pi -p spawn.
	for _, want := range []string{
		"pi.registerTool(",
		`import { Type } from "typebox"`,
		`"-p"`,
		`"--append-system-prompt"`,
		"child_process",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("PI extension missing %q", want)
		}
	}
	// One tool per persona, named by the persona and carrying its body as the prompt.
	for _, name := range []string{"backend-architect", "verifier"} {
		if !strings.Contains(src, `"name": "`+name+`"`) {
			t.Errorf("baked personas missing name %q:\n%s", name, src)
		}
	}
	if !strings.Contains(src, "You design things") {
		t.Errorf("persona body must be baked as the subagent prompt:\n%s", src)
	}
	// The description carries embedded quotes; JSON encoding must escape them so the
	// baked array is valid JS.
	if !strings.Contains(src, `\"how should this be structured\"`) {
		t.Errorf("description quotes not JSON-escaped:\n%s", src)
	}
	// A human-readable label is derived (PI requires one).
	if !strings.Contains(src, `"label": "Backend Architect"`) {
		t.Errorf("expected a derived label:\n%s", src)
	}
}

// The baked extension must be deterministic (byte-stable) so check can verify it.
func TestPIExtensionMirror_Deterministic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "docs/agents/b.md", personaWithQuotes)
	writeFile(t, dir, "docs/agents/a.md", anotherValidPersona)
	m1, _, err := PIExtensionMirror(dir)
	if err != nil {
		t.Fatal(err)
	}
	m2, _, _ := PIExtensionMirror(dir)
	if string(m1.Bytes) != string(m2.Bytes) {
		t.Error("PI extension render is not deterministic")
	}
}

func TestLabel(t *testing.T) {
	cases := map[string]string{
		"backend-architect": "Backend Architect",
		"code_reviewer":     "Code Reviewer",
		"verifier":          "Verifier",
	}
	for in, want := range cases {
		if got := label(in); got != want {
			t.Errorf("label(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestList_ParsesValidPersonasSorted(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "docs/agents/zeta.md", anotherValidPersona)
	writeFile(t, dir, "docs/agents/alpha.md", personaWithQuotes)
	writeFile(t, dir, "docs/agents/README.md", "# not a persona\n")
	ps, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 2 {
		t.Fatalf("got %d personas, want 2", len(ps))
	}
	if ps[0].Name != "backend-architect" || ps[1].Name != "verifier" {
		// alpha.md holds backend-architect; zeta.md holds verifier — sorted by NAME.
		t.Errorf("List not sorted by persona name: %q, %q", ps[0].Name, ps[1].Name)
	}
}

// Guard: os.ReadDir underpins List; ensure a genuinely missing docs/agents is clean.
func TestList_MissingDirIsEmpty(t *testing.T) {
	if ps, err := List(t.TempDir()); err != nil || len(ps) != 0 {
		t.Fatalf("missing docs/agents: ps=%v err=%v", ps, err)
	}
}

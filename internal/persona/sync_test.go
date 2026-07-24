package persona

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestSyncScript_BakesIdempotentWrites(t *testing.T) {
	mirrors := []Mirror{
		{Src: "docs/agents/a.md", Dest: ".claude/agents/a.md", Bytes: []byte("hello <world> & \"quotes\"")},
		{Src: "docs/agents/a.md", Dest: ".codex/agents/a.toml", Bytes: []byte("name = \"a\"\n")},
	}
	s := string(SyncScript(mirrors))

	// Skip-if-exists (idempotent) + base64 decode + mkdir.
	for _, want := range []string{`[ -f "$1" ] && return 0`, "base64 -d", `mkdir -p "$(dirname "$1")"`} {
		if !strings.Contains(s, want) {
			t.Errorf("script missing %q:\n%s", want, s)
		}
	}
	// One write per mirror, with the dest and the base64 of its exact bytes.
	for _, m := range mirrors {
		enc := base64.StdEncoding.EncodeToString(m.Bytes)
		if !strings.Contains(s, "'"+m.Dest+"'") {
			t.Errorf("script missing dest %q", m.Dest)
		}
		if !strings.Contains(s, "'"+enc+"'") {
			t.Errorf("script missing baked bytes for %q", m.Dest)
		}
	}
}

func TestSyncScript_DeterministicAndEmpty(t *testing.T) {
	mirrors := []Mirror{{Dest: ".claude/agents/a.md", Bytes: []byte("x")}}
	a, b := string(SyncScript(mirrors)), string(SyncScript(mirrors))
	if a != b {
		t.Error("SyncScript not deterministic")
	}
	// No mirrors → just the header, no `w ` invocations.
	if strings.Contains(string(SyncScript(nil)), "\nw '") {
		t.Error("empty mirror set should bake no write lines")
	}
}

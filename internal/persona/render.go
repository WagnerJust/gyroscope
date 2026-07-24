package persona

import (
	"path/filepath"
	"strings"
)

// Persona is a canonical persona parsed from docs/agents/. Raw is the original
// file bytes — used verbatim for the Claude mirror, whose format is identical to
// the source. Name/Description/Body are parsed for harnesses whose subagent format
// differs (OpenCode, Codex), so the binary can re-encode rather than byte-copy.
type Persona struct {
	Name        string
	Description string   // raw text after `description:` (verbatim; may contain quotes)
	Tools       []string // the `tools:` allowlist (Claude tool names), nil when the field is absent
	Body        string   // everything after the closing frontmatter `---`
	Raw         []byte
}

// parse reads a persona file. ok is false when it is not a valid persona — no
// leading YAML frontmatter, or no non-empty `name:` — because such a file cannot
// be registered as a subagent and is skipped. Only the `key: value` subset of YAML
// personas use is read; no YAML dependency is pulled in (dependency-light).
func parse(b []byte) (Persona, bool) {
	s := string(b)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return Persona{}, false
	}
	lines := strings.Split(s, "\n")
	var name, desc string
	var tools []string
	end := -1
	for i := 1; i < len(lines); i++ {
		t := strings.TrimRight(lines[i], "\r")
		if t == "---" {
			end = i
			break
		}
		key, val, ok := strings.Cut(t, ":") // first colon only, so a `:` in the value is kept
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "name":
			name = strings.TrimSpace(val)
		case "description":
			desc = strings.TrimSpace(val)
		case "tools":
			tools = splitList(val)
		}
	}
	if end == -1 || name == "" {
		return Persona{}, false
	}
	return Persona{
		Name:        name,
		Description: desc,
		Tools:       tools,
		Body:        strings.Join(lines[end+1:], "\n"),
		Raw:         b,
	}, true
}

// splitList parses a comma-separated allowlist ("Read, Grep, Bash") into trimmed,
// non-empty entries. Returns nil for an empty value so an absent/blank `tools:` is
// distinguishable from a real allowlist.
func splitList(val string) []string {
	var out []string
	for _, p := range strings.Split(val, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// Target is one harness's subagent registry: where a persona registers, and how
// its canonical docs/agents/ form is re-encoded into that harness's format. The
// binary re-encodes gyroscope-OWNED mirror files deterministically — it does not
// template user content or fill interview data (see ADR 0014, extending 0003/0010).
type Target struct {
	ID     string
	dest   func(name string) string
	render func(Persona) []byte
}

// Claude registers subagents from .claude/agents/*.md in the SAME markdown format
// as docs/agents/, so its mirror is a byte-copy (ADR 0010).
var Claude = Target{
	ID:     "claude",
	dest:   func(name string) string { return filepath.Join(".claude", "agents", name+".md") },
	render: func(p Persona) []byte { return p.Raw },
}

// OpenCode registers subagents from .opencode/agents/*.md — markdown with YAML
// frontmatter and the body as the system prompt, like Claude, but it needs
// `mode: subagent` and does not use Claude's `name`/`tools`/`model` keys (agent id
// comes from the filename; `tools` is deprecated for `permission`). So the body is
// kept verbatim and only the frontmatter is re-encoded.
var OpenCode = Target{
	ID:     "opencode",
	dest:   func(name string) string { return filepath.Join(".opencode", "agents", name+".md") },
	render: renderOpenCode,
}

// Codex registers subagents from .codex/agents/*.toml — TOML, where the system
// prompt is a `developer_instructions` field rather than the file body. So the
// whole persona is re-encoded from markdown into TOML.
var Codex = Target{
	ID:     "codex",
	dest:   func(name string) string { return filepath.Join(".codex", "agents", name+".toml") },
	render: renderCodex,
}

func renderOpenCode(p Persona) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("description: ")
	b.WriteString(p.Description)
	b.WriteByte('\n')
	b.WriteString("mode: subagent\n")
	b.WriteString(openCodePermissionBlock(p.Tools))
	b.WriteString("---\n")
	b.WriteString(p.Body)
	return []byte(b.String())
}

func renderCodex(p Persona) []byte {
	var b strings.Builder
	b.WriteString("# gyroscope-managed mirror of docs/agents/")
	b.WriteString(p.Name)
	b.WriteString(".md — do not edit by hand.\n")
	b.WriteString("name = ")
	b.WriteString(tomlBasic(p.Name))
	b.WriteByte('\n')
	b.WriteString("description = ")
	b.WriteString(tomlBasic(p.Description))
	b.WriteByte('\n')
	if mode := codexSandboxMode(p.Tools); mode != "" {
		b.WriteString("sandbox_mode = ")
		b.WriteString(tomlBasic(mode))
		b.WriteByte('\n')
	}
	b.WriteString("developer_instructions = ")
	b.WriteString(tomlMultiline(p.Body))
	b.WriteByte('\n')
	return []byte(b.String())
}

// tomlBasic encodes s as a single-line TOML basic string, escaping the characters
// TOML disallows raw. Descriptions carry embedded double quotes (e.g. Use for "x"),
// so escaping is not optional.
func tomlBasic(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"\n", `\n`,
		"\r", `\r`,
		"\t", `\t`,
	)
	return `"` + r.Replace(s) + `"`
}

// tomlMultiline encodes s as a multi-line TOML basic string. Backslashes are
// escaped first, then any `"""` that would prematurely close the string is
// neutralized. A newline immediately after the opening delimiter is trimmed by
// TOML, so one is emitted to keep the body on its own line.
func tomlMultiline(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"""`, `\"\"\"`)
	return "\"\"\"\n" + s + "\"\"\""
}

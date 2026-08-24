// Package command re-encodes a repo's canonical slash commands from docs/commands/
// into each harness's native command format, so a command authored once is
// dispatchable in every declared harness. It mirrors internal/persona: one neutral
// source, a per-harness Target that relocates + re-encodes, verified byte-for-byte
// by `gyroscope check`. The command BODY (prompt, with $ARGUMENTS / !`cmd` / @file)
// is portable across Claude, OpenCode, and PI, so only the YAML frontmatter keys and
// file location change (see ADR 0019). The canonical form is the Claude command
// format — the most-featured superset.
package command

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/WagnerJust/gyroscope/internal/persona"
)

// Command is a parsed canonical command: its name (from the source filename), the
// ordered frontmatter key/value pairs, and the prompt body. Raw is the original
// bytes, used verbatim for the Claude target whose format is identical.
type Command struct {
	Name  string
	Front []kv
	Body  string
	Raw   []byte
}

type kv struct{ key, val string }

// parse reads a command file named name (without the .md extension). Frontmatter is
// OPTIONAL — unlike a persona, any markdown file is a valid command (a bare prompt
// body with no frontmatter is fine). Only single-line `key: value` frontmatter is
// read; a rare multi-line value (e.g. an `allowed-tools` YAML list) is dropped from
// the non-Claude targets anyway, and preserved verbatim for Claude via Raw.
func parse(name string, b []byte) Command {
	s := string(b)
	c := Command{Name: name, Raw: b}
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		c.Body = s // no frontmatter: the whole file is the prompt
		return c
	}
	lines := strings.Split(s, "\n")
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			end = i
			break
		}
		key, val, ok := strings.Cut(lines[i], ":")
		if !ok {
			continue
		}
		c.Front = append(c.Front, kv{strings.TrimSpace(key), strings.TrimSpace(val)})
	}
	if end == -1 {
		// Opening marker with no close: treat the whole file as the body (don't guess).
		return Command{Name: name, Body: s, Raw: b}
	}
	c.Body = strings.Join(lines[end+1:], "\n")
	return c
}

// front returns the verbatim value of a frontmatter key, or "" if absent.
func (c Command) front(key string) string {
	for _, e := range c.Front {
		if e.key == key {
			return e.val
		}
	}
	return ""
}

// Target is one harness's command registry: where a command lands and how it is
// re-encoded. keep lists the frontmatter keys that harness supports (in canonical
// order); the body is always kept verbatim.
type Target struct {
	ID     string
	dest   func(name string) string
	render func(Command) []byte
}

// Claude registers commands from .claude/commands/*.md in the SAME format as the
// canonical source, so its target is a verbatim copy.
var Claude = Target{
	ID:     "claude",
	dest:   func(n string) string { return filepath.Join(".claude", "commands", n+".md") },
	render: func(c Command) []byte { return c.Raw },
}

// OpenCode reads .opencode/commands/*.md. It supports a `description` (its other
// frontmatter — agent/model — is format-specific), so only description is carried;
// the body's $ARGUMENTS / !`cmd` / @file syntax is native.
var OpenCode = Target{
	ID:     "opencode",
	dest:   func(n string) string { return filepath.Join(".opencode", "commands", n+".md") },
	render: renderKeep([]string{"description"}),
}

// PI reads prompt templates from .pi/prompts/*.md, supporting `description` and
// `argument-hint`, with native $ARGUMENTS support.
var PI = Target{
	ID:     "pi",
	dest:   func(n string) string { return filepath.Join(".pi", "prompts", n+".md") },
	render: renderKeep([]string{"description", "argument-hint"}),
}

// renderKeep builds a renderer that emits a frontmatter block with only the kept
// keys (verbatim values, canonical order) followed by the verbatim body. When no
// kept key is present, it emits the body alone (no empty `---` block).
func renderKeep(keep []string) func(Command) []byte {
	return func(c Command) []byte {
		var fm strings.Builder
		for _, k := range keep {
			if v := c.front(k); v != "" {
				fm.WriteString(k)
				fm.WriteString(": ")
				fm.WriteString(v)
				fm.WriteByte('\n')
			}
		}
		if fm.Len() == 0 {
			return []byte(c.Body)
		}
		return []byte("---\n" + fm.String() + "---\n" + c.Body)
	}
}

// PlanFor lists the command mirrors gyroscope would write for target t: one per
// *.md under docs/commands/. Reuses persona.Mirror as the generic "generated file"
// record (Src/Dest/Bytes) so both artifact types flow through the same write, check,
// and sync-bake paths. A missing docs/commands/ yields an empty plan, not an error.
func PlanFor(repoDir string, t Target) ([]persona.Mirror, error) {
	dir := filepath.Join(repoDir, "docs", "commands")
	entries, err := readMarkdown(dir)
	if err != nil {
		return nil, err
	}
	var mirrors []persona.Mirror
	for _, e := range entries {
		name := strings.TrimSuffix(e.name, ".md")
		c := parse(name, e.bytes)
		mirrors = append(mirrors, persona.Mirror{
			Src:   filepath.Join("docs", "commands", e.name),
			Dest:  t.dest(name),
			Bytes: t.render(c),
		})
	}
	return mirrors, nil
}

type mdFile struct {
	name  string
	bytes []byte
}

// readMarkdown returns the *.md files directly under dir, in name order (os.ReadDir
// is sorted, so results are deterministic). A missing dir yields nil, not an error.
func readMarkdown(dir string) ([]mdFile, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []mdFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, mdFile{e.Name(), b})
	}
	return out, nil
}

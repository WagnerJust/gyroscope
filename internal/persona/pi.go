package persona

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// piPersonaExtensionPath is the auto-discovered, project-local PI extension that
// registers each persona as a dispatchable subagent. It is separate from the
// context-injection extension (gyroscope-context.ts) so the two concerns stay
// independent; PI loads every extension under .pi/extensions/.
const piPersonaExtensionPath = ".pi/extensions/gyroscope-personas.ts"

// piPersona is the baked metadata for one registered subagent. gyroscope parses the
// canonical persona once and bakes this JSON into the extension, so the extension
// never re-parses YAML (no dual-implementation drift) — the single-parser rule that
// also shapes the other harness mirrors (ADR 0010/0014).
type piPersona struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

// List returns every valid persona under repoDir/docs/agents/, parsed, in
// filename order (os.ReadDir is sorted, so the result is deterministic). A missing
// docs/agents/ yields an empty list, not an error.
func List(repoDir string) ([]Persona, error) {
	agentsDir := filepath.Join(repoDir, "docs", "agents")
	entries, err := os.ReadDir(agentsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Persona
	for _, e := range entries {
		if e.IsDir() || e.Name() == "README.md" || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(agentsDir, e.Name()))
		if err != nil {
			return nil, err
		}
		if p, ok := parse(b); ok {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// PIExtensionMirror returns the single mirror file that registers repoDir's personas
// as PI subagents, and ok=false when there are no personas (so nothing is written for
// a persona-less repo). It rides the ordinary Mirror pipeline — written by Write,
// verified byte-for-byte by check — because, like the file mirrors, it is a
// gyroscope-owned generated artifact. PI has no native subagent registry, so parity
// is this runtime extension rather than a file copy (ADR 0006).
func PIExtensionMirror(repoDir string) (Mirror, bool, error) {
	ps, err := List(repoDir)
	if err != nil {
		return Mirror{}, false, err
	}
	if len(ps) == 0 {
		return Mirror{}, false, nil
	}
	content, err := renderPIExtension(ps)
	if err != nil {
		return Mirror{}, false, err
	}
	return Mirror{
		Src:   filepath.Join("docs", "agents"),
		Dest:  filepath.FromSlash(piPersonaExtensionPath),
		Bytes: content,
	}, true, nil
}

// renderPIExtension bakes the persona metadata into the managed extension. The
// system prompt of each subagent is the persona BODY (its instructions), delegated
// to a fresh headless `pi -p` run restricted to that prompt.
func renderPIExtension(ps []Persona) ([]byte, error) {
	metas := make([]piPersona, 0, len(ps))
	for _, p := range ps {
		metas = append(metas, piPersona{
			Name:        p.Name,
			Label:       label(p.Name),
			Description: p.Description,
			Prompt:      strings.TrimLeft(p.Body, "\n"),
		})
	}
	// Encode with HTML escaping off so a `<`/`>` in a persona body stays literal;
	// the result is still valid JS (a JSON array literal).
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metas); err != nil {
		return nil, err
	}
	return fmt.Appendf(nil, piExtensionTemplate, strings.TrimRight(buf.String(), "\n")), nil
}

// label turns a persona slug into a human-readable tool label ("backend-architect"
// → "Backend Architect"), which PI's ToolDefinition requires.
func label(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// piExtensionTemplate is the managed PI persona extension. %s is the JSON array of
// baked persona metadata. Each persona becomes a subagent tool whose execute()
// spawns a fresh headless `pi -p` with the persona body as an appended system prompt
// and returns its printed response. Tool restriction (the persona `tools:` field) is
// intentionally NOT mapped to `--tools`: Claude tool names do not map cleanly onto
// PI's, and a wrong name would fail the run — so the subagent runs with PI's default
// toolset. That is a documented parity gap, not a silent one.
const piExtensionTemplate = `// gyroscope-personas.ts — managed by gyroscope; do not edit by hand.
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { Type } from "typebox";
import { spawn } from "node:child_process";

interface Persona { name: string; label: string; description: string; prompt: string; }

const PERSONAS: Persona[] = %s;

// piInvocation resolves how to spawn a headless pi: re-invoke this running pi script
// when possible, else fall back to ` + "`pi`" + ` on PATH (mirrors PI's own subagent example).
function piInvocation(): { command: string; args: string[] } {
	const argv1 = process.argv[1];
	if (argv1 && /(^|[\\/])pi(\.[cm]?js)?$/.test(argv1)) {
		return { command: process.execPath, args: [argv1] };
	}
	return { command: "pi", args: [] };
}

export default function (pi: ExtensionAPI) {
	for (const persona of PERSONAS) {
		pi.registerTool({
			name: persona.name,
			label: persona.label,
			description: persona.description,
			parameters: Type.Object({
				task: Type.String({ description: "The task to delegate to this persona." }),
			}),
			async execute(_toolCallId, params, signal) {
				const inv = piInvocation();
				const args = [
					...inv.args,
					"-p",
					"--no-session",
					"--append-system-prompt",
					persona.prompt,
					"Task: " + params.task,
				];
				const text = await new Promise<string>((resolve, reject) => {
					const proc = spawn(inv.command, args, {
						shell: false,
						stdio: ["ignore", "pipe", "pipe"],
						signal,
					});
					let out = "";
					let err = "";
					proc.stdout.on("data", (d) => (out += d));
					proc.stderr.on("data", (d) => (err += d));
					proc.on("error", reject);
					proc.on("close", (code) =>
						code === 0 ? resolve(out.trim()) : reject(new Error(err.trim() || "pi exited " + code)),
					);
				});
				return { content: [{ type: "text", text }], details: {} };
			},
		});
	}
}
`

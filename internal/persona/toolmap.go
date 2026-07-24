package persona

import "strings"

// This file carries a persona's `tools:` allowlist (Claude tool names) through to
// each harness's own tool-restriction model, so a restricted persona stays
// restricted when dispatched. The models differ: OpenCode has a per-tool permission
// map (faithful), PI has a `--tools` allowlist (faithful for the tools it has), and
// Codex has only a coarse sandbox mode (approximated). Mappings are verified against
// each tool's docs/source; see ADR 0014/0015.

// claudeToOpenCode maps Claude tool names to OpenCode permission keys, in a fixed
// order so the emitted permission block is deterministic. All eight map cleanly.
var claudeToOpenCode = []struct{ claude, opencode string }{
	{"Read", "read"},
	{"Write", "write"},
	{"Edit", "edit"},
	{"Grep", "grep"},
	{"Glob", "glob"},
	{"Bash", "bash"},
	{"WebFetch", "webfetch"},
	{"WebSearch", "websearch"},
}

// claudeToPI maps Claude tool names to PI's built-in tool names. WebFetch/WebSearch
// have NO PI built-in equivalent, so they are absent (dropped, never passed to
// `--tools` where an unknown name could fail the run).
var claudeToPI = map[string]string{
	"Read":  "read",
	"Write": "write",
	"Edit":  "edit",
	"Grep":  "grep",
	"Glob":  "find", // PI's glob-based file search; closest equivalent
	"Bash":  "bash",
}

// openCodePermissionBlock renders an OpenCode agent `permission:` block that allows
// exactly the persona's tools and denies the rest of the mappable set (a Claude
// allowlist expressed as OpenCode's deny-capable permission map — `deny` is a hard
// block). Returns "" when the persona declares no allowlist (leave it unrestricted).
func openCodePermissionBlock(tools []string) string {
	if len(tools) == 0 {
		return ""
	}
	allowed := make(map[string]bool, len(tools))
	for _, t := range tools {
		allowed[t] = true
	}
	var b strings.Builder
	b.WriteString("permission:\n")
	for _, m := range claudeToOpenCode {
		v := "deny"
		if allowed[m.claude] {
			v = "allow"
		}
		b.WriteString("  ")
		b.WriteString(m.opencode)
		b.WriteString(": ")
		b.WriteString(v)
		b.WriteByte('\n')
	}
	return b.String()
}

// codexSandboxMode returns "read-only" when the persona's allowlist is read-only
// (only Read/Grep/Glob — no mutating or network tool), else "" to leave Codex's
// default untouched. Codex has no per-tool allowlist, only the coarse sandbox mode,
// so this only ever TIGHTENS (a read-only persona), never loosens (never forces a
// broader mode over a stricter default).
func codexSandboxMode(tools []string) string {
	if len(tools) == 0 {
		return ""
	}
	for _, t := range tools {
		switch t {
		case "Read", "Grep", "Glob":
		default:
			return ""
		}
	}
	return "read-only"
}

// piTools maps a persona's Claude allowlist to PI built-in tool names, dropping any
// with no PI equivalent (WebFetch/WebSearch) and de-duplicating, preserving order.
// Only verified names are returned, so the result is always safe to pass to
// `--tools`.
func piTools(claude []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, c := range claude {
		if pi, ok := claudeToPI[c]; ok && !seen[pi] {
			seen[pi] = true
			out = append(out, pi)
		}
	}
	return out
}

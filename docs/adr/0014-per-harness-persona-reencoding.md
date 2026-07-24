# ADR 0014: the binary re-encodes personas into each harness's subagent format

- **Status:** accepted
- **Date:** 2026-07-23

## Context

ADR 0010 registers a repo's `docs/agents/` personas as Claude subagents by
**byte-copying** each into `.claude/agents/<name>.md`. Byte-copy was chosen
deliberately: the two formats are identical (markdown + YAML frontmatter + body),
so no rendering is needed, and ADR 0003 forbids a templating engine in the binary
(content comes from the skill, not from the binary substituting variables).

Making other harnesses genuinely "ready" — not just able to read the `AGENTS.md`
hub, but able to **dispatch the same personas as subagents** — reopened that
choice. OpenCode and Codex both have native, project-committable subagent
registries, but their formats differ from `docs/agents/`:

- **OpenCode** — `.opencode/agents/<name>.md`: markdown + frontmatter + body like
  Claude, but the agent id comes from the filename, it needs `mode: subagent`, and
  Claude's `name`/`tools`/`model` keys do not apply (`tools` is deprecated for
  `permission`). So the body copies verbatim but the frontmatter must be rewritten.
- **Codex** — `.codex/agents/<name>.toml`: TOML, where the system prompt is a
  `developer_instructions` field, not the file body. So the whole persona must be
  re-encoded from markdown into TOML.

A pure byte-copy can no longer serve. The question ADR 0003/0010 left implicit had
to be settled: **who re-encodes?** The genuine alternatives:

- **The skill authors the extra files.** The `/gyroscope` skill already authors
  persona *content*, so it could write `.opencode/agents/` and `.codex/agents/`
  during the interview, keeping the binary a pure byte-copier. Rejected: `check`
  must still parse to verify them, so the binary gains the parsing burden anyway;
  convergence would need a skill re-run rather than the deterministic `check --fix`
  loop; and mirror ownership would be split across the binary (Claude) and the skill
  (others) for no principled reason.
- **Ship only settings, not personas.** Rejected: it abandons the actual goal
  (dispatchable personas) for the two harnesses that support them.
- **The binary re-encodes.** Chosen.

## Decision

The binary re-encodes each canonical persona into every enabled harness's subagent
format. A `persona.Target` names a harness's registry (`dest` path + `render`
function); `PlanFor(repoDir, target)` reads `docs/agents/`, parses each persona once,
and emits per target:

- **Claude** → `.claude/agents/<name>.md`, byte-copy (unchanged from ADR 0010).
- **OpenCode** → `.opencode/agents/<name>.md`, frontmatter rewritten to
  `description` + `mode: subagent`, body verbatim.
- **Codex** → `.codex/agents/<name>.toml`, emitted as TOML with the body as a
  `developer_instructions` multi-line string (quotes/backslashes/`"""` escaped).

This is **not** the templating engine ADR 0003 forbids. It performs no variable
substitution and consumes no interview data; it mechanically re-encodes
gyroscope-OWNED mirror files whose canonical source is `docs/agents/`. That is the
same category as generating the Claude mirror — only the target grammar changed. The
mirrors stay gyroscope-owned and overwrite-on-drift (the existing atomic,
non-`WriteGuarded` mirror write), and `check` verifies each byte-for-byte, so any
hand-edit is caught and `check --fix` re-encodes.

Each non-Claude registry is gated by an **opt-in** `enforce.opencode` /
`enforce.codex` toggle (default false), consistent with ADR 0002/0006 — enforcement
is opt-in per harness — and, like the Claude mirror, only runs when `personas == on`.

## Consequences

- OpenCode and Codex reach persona parity with Claude: the same authored personas
  dispatch as subagents in all three, from one canonical source.
- The binary now owns three encoders instead of one copy. The re-encode logic is
  small, pure, and unit-tested; the ADR 0003 boundary is redrawn precisely: "no
  templating of user content" holds, "no format re-encoding of gyroscope's own
  files" was never the rule (the Claude mirror already generated files).
- PI is deliberately excluded: it has no native subagent registry (ADR 0006), so its
  parity is an extension-built runtime, not a file mirror — a separate decision.
- Codex's `commit_attribution` setting (its attribution kill-switch) is wired
  separately under the same `enforce.codex` gate; OpenCode has no attribution setting,
  so it relies on the cross-harness hub directive alone.

# ADR 0019: cross-harness slash commands via a neutral canonical source

- **Status:** accepted
- **Date:** 2026-08-24

## Context

Slash commands authored for one harness (typically Claude's `.claude/commands/*.md`)
do not work in others, so a team that shares a repo across Claude Code, OpenCode, and
PI must re-author each command per tool. gyroscope already solves the analogous
problem for personas — one canonical `docs/agents/` source re-encoded into each
harness's subagent registry (ADR 0014/0015). Commands are the same shape of problem.

Research (2026-08-24) confirmed the three harnesses' command mechanisms are close:
all are markdown + YAML frontmatter + a prompt body, and the **body syntax is
portable** — `$ARGUMENTS`/`$1`, `` !`cmd` `` shell injection, and `@file` references
work in Claude, OpenCode (`.opencode/commands/*.md`), and PI (`.pi/prompts/*.md`).
Only the frontmatter keys and the file location differ.

## Decision

Add slash commands as a second re-encoded artifact type, built on the persona
pattern (an `internal/command` package that mirrors `internal/persona`):

- **Neutral canonical source:** `docs/commands/<name>.md`, authored in the Claude
  command format (the most-featured superset). Frontmatter is optional — any markdown
  file is a valid command (a bare prompt body is fine). The command name is the
  filename.
- **Per-harness re-encode** via a `command.Target` (dest + render), body always
  verbatim (portable), frontmatter reduced to what each harness supports:
  - Claude → `.claude/commands/<name>.md`, verbatim (canonical == Claude format).
  - OpenCode → `.opencode/commands/<name>.md`, keep `description`.
  - PI → `.pi/prompts/<name>.md`, keep `description` + `argument-hint`.
  Claude-only keys (`allowed-tools`, `model`, and `argument-hint` for OpenCode) are
  dropped rather than mistranslated — the same conservative choice as the persona
  `tools:` mapping, since a harness-specific `model` value would not be valid
  elsewhere.
- **Gating:** per enabled harness (`enforce.{claude,opencode,pi}`), presence-driven —
  a repo with `docs/commands/*.md` gets them installed for its enabled harnesses; a
  missing `docs/commands/` is a no-op. Commands are **independent of the personas
  state** (a commands-only repo works with personas off).
- **Verification & reconstruction:** the generated command files are gyroscope-owned
  mirrors — `check` verifies each byte-for-byte, `check --fix` re-encodes, and they
  ride the same generated-file list (`planMirrors`) as persona mirrors, so the
  `gyroscope-sync.sh` reconstruction (ADR 0018) bakes them too. The binary re-encodes
  gyroscope-owned output (ADR 0014's boundary); the skill authors the canonical
  command bodies.

## Consequences

- A command authored once in `docs/commands/` is dispatchable as `/name` in every
  declared harness, from one source — closing the per-tool re-authoring gap and
  extending the "works without Claude" story to commands.
- `planMirrors` becomes the single "all gyroscope-owned generated files" list
  (personas + commands), so init/check/sync agree on the full set without each
  learning about commands separately.
- Dropped frontmatter is a deliberate fidelity trade (safety over mistranslation);
  richer per-harness frontmatter mapping can be added later without changing the
  shape, exactly like the persona tool-map follow-up.
- The `docs/commands/` hub route + scaffold and the skill's adopt-existing-commands
  flow are follow-ups (skill/docs phase), not part of this generation mechanism.

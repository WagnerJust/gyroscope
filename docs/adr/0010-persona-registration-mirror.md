# ADR 0010: Persona registration mirror — docs/agents → .claude/agents

- **Status:** accepted
- **Date:** 2026-07-10

## Context

A repo's `docs/agents/*.md` personas are authored (by the `/gyroscope` skill) in
valid Claude Code subagent format: YAML frontmatter (`name`/`description`/`tools`/
`model`) plus a system-prompt body. But they live in `docs/agents/`, where the hub
routes to them for humans and doc-reading agents — and Claude Code does **not** scan
`docs/agents/` for subagents. Claude registers subagents only from `.claude/agents/`
(project) or `~/.claude/agents/` (user).

The result (finding #2) is an asymmetry: a persona can be perfectly authored and
hub-routed yet never become a *dispatchable* agent type. When the model looks for a
specialized agent, it finds only the ones Claude actually registered — the caveman
`cavecrew` plugin agents — and reaches for those instead of the repo's own personas.
The docs said "here are your personas"; the registry said "there are none." cavecrew
won by being the only thing registered.

The genuine alternatives for closing the asymmetry:

- **Move the personas into `.claude/agents/` outright.** Rejected: it surrenders
  `docs/agents/` as the canonical, hub-routed home, couples the personas to one
  harness's directory layout, and makes them invisible to every non-Claude tool the
  hub routes.
- **Symlink `.claude/agents/<name>.md` → `docs/agents/<file>`.** Rejected: symlinks
  are fragile across clones, archives, and Windows; a broken link registers nothing
  and fails silently; and there is no natural point to detect the breakage.
- **Copy (mirror) each persona into `.claude/agents/`.** Chosen — see below.

## Decision

Keep `docs/agents/` as the canonical, hub-routed source, and **mirror** each valid
persona into `.claude/agents/<name>.md`, byte-for-byte, so Claude registers it.

- **Copy, not symlink.** A plain byte copy is robust across clones and operating
  systems, and any divergence between the source and its mirror is *detectable* —
  `gyroscope check` re-derives the expected bytes and flags a missing or drifted
  mirror as nonconformance; `check --fix` (and `init --apply`) re-mirror to converge.
  Symlink fragility is traded for a copy that can drift, and the drift is caught.
- **Valid persona only.** A file qualifies iff it is a `docs/agents/*.md` other than
  `README.md` whose frontmatter carries a `name:`. The dest filename is that `name:`,
  so `.claude/agents/<name>.md` matches the registered subagent type even when the
  source file is named differently.
- **gyroscope owns the mirror files.** The persona-named files under `.claude/agents/`
  are a *generated mirror*, like the SessionStart hook is idempotently re-applied.
  So the writer overwrites on drift — the one deliberate exception to `WriteGuarded`'s
  O_EXCL refuse-to-clobber guarantee (a re-mirror would otherwise be refused). The
  write is atomic (temp file + rename) and only ever touches `.claude/agents/<name>.md`
  — never `.claude/settings.json` or any other `.claude/` content.
- **Gated.** The mirror runs only when the persona spoke is actually wired
  (`spokes.personas == on`) AND the Claude enforcement adapter is enabled
  (`enforce.claude`). An unknown/skipped/off persona spoke mirrors nothing, and a
  PI-only repo (Claude adapter off) mirrors nothing — PI has its own agent mechanism.

The mirror is a byte copy, not a render. This is a deliberate **scope shift** for the
binary: it now **copies persona files for registration**, but it still never
**authors** persona content — the `/gyroscope` skill authors personas; the binary
mirrors their bytes. That is consistent with the standing invariant "the binary copies
bytes, it does not render": no templating engine is added, and the copy stays
deterministic and CI-safe.

## Consequences

- Personas authored under `docs/agents/` become dispatchable Claude subagents after
  `init --apply`, so the model reaches for the repo's own personas rather than
  cavecrew. The docs-vs-registered asymmetry is closed.
- A new kind of gyroscope-owned file exists: a generated mirror under `.claude/agents/`
  that is overwritten (not O_EXCL-guarded) and verified by `check`. It is the single
  place the binary's "never clobber user work" guarantee is deliberately relaxed,
  justified because the file is gyroscope's own output, not the user's.
- The binary's responsibility widens from "structure + hook" to "structure + hook +
  persona registration (byte copy)", while the binary/skill split holds: authoring
  stays with the skill.
- Drift is now possible (a mirror can be hand-edited or a source can change) but is
  caught: `check` reports a missing or differing mirror, and `--fix` re-converges.
- Scope excluded: `README.md` and frontmatter-less files never register; the user-scope
  `~/.claude/agents/` is untouched (project scope only); and PI registration is out of
  scope (its own mechanism, a future adapter).

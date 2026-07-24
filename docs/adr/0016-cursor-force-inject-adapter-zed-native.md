# ADR 0016: Cursor gets a force-inject adapter; Zed needs none (native read)

- **Status:** accepted
- **Date:** 2026-07-24

## Context

gyroscope's enforcement registry (ADR 0002) installs, per harness, a mechanism that
makes agents actually read the hub — Claude's SessionStart hook, PI's `session_start`
extension. Cursor and Zed had only the passive *pointer* files gyroscope writes
(`.cursorrules`, `.rules`): a line telling the agent to read `AGENTS.md`, with no
guarantee it is loaded or followed. Whether a real force-inject mechanism exists for
each was an open question (the deferred items said: build it, or if the tool only
natively reads a file, document that there is nothing to enforce). Verified against
each tool's official docs (2026-07-24):

- **Cursor** has a genuine force-inject primitive: a project rule under
  `.cursor/rules/*.mdc` with `alwaysApply: true` is injected into *every* model
  request regardless of file context. It is git-committable. Cursor also reads
  `AGENTS.md` natively, but that is passive; the always-apply rule is the enforced
  analog. (Cursor's legacy `.cursorrules` is deprecated and, per secondary reports,
  may be ignored in agent mode — so the pointer alone is not reliable there.)
- **Zed** has *no* hook or always-apply flag. Its Agent auto-reads a single project
  instruction file — the first match of a fixed precedence list (`.rules` wins over
  `AGENTS.md`) — on *every* session, automatically. So the passive `.rules` pointer
  gyroscope already writes *is* the mechanism: it is structurally guaranteed to load.
  There is nothing to force-inject.

## Decision

- **Build a Cursor enforcement adapter.** `enforce.Cursor` writes
  `.cursor/rules/gyroscope-hub.mdc` with `alwaysApply: true` and a body matching the
  pointer directive ("Before doing anything else, read AGENTS.md and follow its
  routes"). It is opt-in via `enforce.cursor` (default false), like PI (ADR 0002/0006
  — enforcement is opt-in per harness), and rides the same adapter interface (init
  applies it, `check` verifies it, `check --fix` restores it).
  - The rule is **static** — a Cursor rule is injected text and cannot cat live
    files, so it force-injects the *directive to read the hub*, not the hub's
    contents. The hub's routes carry the agent to the state/resume spokes from there.
    This is the ceiling of what Cursor's rule mechanism allows; it does not vary with
    the injected spoke set, so the adapter ignores `paths`.
  - The passive `.cursorrules` pointer is **kept** for backward compatibility; the
    adapter supersedes it for modern agent-mode Cursor without removing it.

- **Do not build a Zed adapter.** Zed natively auto-reads `.rules` every session, so
  the existing pointer is the enforcement — there is no stronger primitive to install.
  Documented here as the deliberate no-op outcome. Two caveats hold: Zed is
  first-match-only (keep `.rules` pointing at the hub, since Zed won't also read
  `AGENTS.md` when `.rules` is present), and loading ≠ obeying (the ceiling for any
  instruction-file harness).

## Consequences

- Cursor joins Claude and PI as a harness with active enforcement, not just a
  pointer — the always-apply rule is the closest Cursor analog to a session hook.
- The enforcement matrix is now: active adapter for Claude / PI / Cursor; native
  auto-read for Zed / OpenCode / Codex (nothing to install); passive pointer for the
  remaining doc-target tools.
- Enforcement is still "context injection, not a hard gate" everywhere — an
  always-apply rule maximizes the odds the hub is read, but no harness offers a true
  gate. That limit is inherent, not a gyroscope gap.

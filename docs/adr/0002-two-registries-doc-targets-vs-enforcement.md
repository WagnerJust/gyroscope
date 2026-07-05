# ADR 0002: Two separate registries — doc targets vs enforcement adapters

- **Status:** accepted
- **Date:** 2026-07-04

## Context

gyroscope has to do two different things to two different sets of tools:

1. Drop a one-line **pointer file** (`CLAUDE.md`, `GEMINI.md`, …) for every tool
   that reads a repo-level instruction file. This set is *large* and mostly
   passive — the tool just needs to be told "read AGENTS.md".
2. Install an **enforcement hook** for the tools that can actually run one, so
   agents are *made* to read the hub rather than choosing to. This set is *small*
   — today only Claude Code (via `SessionStart`); a PI adapter comes later.

The two sets do not line up: many tools take a pointer but cannot be enforced,
and enforcement mechanisms are harness-specific. The genuine alternative was a
single "tool" abstraction with an optional enforcement capability. That would
have coupled a fast-growing, uniform list (pointers) to a slow-growing, wildly
non-uniform one (hooks), forcing every pointer-only tool to carry a null
enforcement field and every new harness to touch the pointer code.

## Decision

Model them as two independent registries:

- `internal/target` — the **doc-target registry**: `{ID, Name, Path}` per tool,
  each writing the same canonical routing line.
- `internal/enforce` — the **enforcement-adapter registry**: an adapter per
  harness that can install a mechanism (the Claude adapter merges a `SessionStart`
  hook into `.claude/settings.json`).

Enforcement is opt-in per harness and lives behind an adapter boundary, so a new
harness drops in without touching the doc writer.

## Consequences

- Adding a pointer target is a one-line registry entry; adding a harness is a new
  adapter — neither disturbs the other.
- The two registries can (and do) grow at very different rates without pressure to
  unify them.
- There is a small seam to keep honest: a tool that is *both* a doc target and an
  enforceable harness (Claude) appears in both registries under the same `ID`.
  That duplication is intentional — the shared ID is the only link, and each
  registry stays single-responsibility.

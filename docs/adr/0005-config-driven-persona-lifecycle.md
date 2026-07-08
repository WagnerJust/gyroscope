# ADR 0005: Config-driven persona lifecycle

- **Status:** accepted
- **Date:** 2026-07-07

## Context

The `docs/agents/` personas spoke shipped as an empty blessed home; populating it
was a manual per-repo chore, and nothing prompted a session to do it. We wanted a
caveman-analog ergonomic — offered, not forced — but concrete and stateful: a
fresh session should proactively offer to wire personas from a template library on
the user's machine, and the decision should persist.

Two constraints shaped the design. First, gyroscope's SessionStart hook is a pure
`cat` with no runtime dependency on the binary (a run-once scaffolder leaves a
self-sufficient repo). Second, the binary owns structure and config; the skill
owns content, and personas are content.

Genuine alternatives considered: (a) a hook that invokes `gyroscope
session-context` to compute the nudge in Go — cleaner state model, but couples
every session to the binary being installed; (b) a tracked `docs/agents/PENDING.md`
sentinel cat by the hook — but that duplicates the JSON state as prose and needs
file-sync logic.

## Decision

`spokes.personas` becomes a four-state enum — `unknown | on | skipped | off` —
back-compatible with the legacy bool (`true`→`unknown`, `false`→`off`). Only
`unknown` prompts action.

The hub carries a standing directive: when `spokes.personas` is `unknown`, ask the
user to wire personas (`/gyroscope`) or skip. The SessionStart hook also cats
`gyroscope.json`, so the live state is in session context and the model can act on
the directive — the hook stays a pure `cat`. `gyroscope agents set <state>` records
decisions. The `/gyroscope` skill does all persona content work; the binary never
reads or customizes persona files. The template directory is per-machine and asked
at wire-time, never persisted — it must not leak a personal path into a shared repo.

Both rejected alternatives were declined for the reasons in Context: no new runtime
dependency, and no second source of truth.

## Consequences

- The nudge rides on the existing cat-only hook plus a standing hub rule, not a
  `gyroscope`-invoking hook — no new runtime dependency.
- State lives only in `gyroscope.json`; no sentinel file to keep in sync.
- `check` gains conformance for the new states (directive present when enabled;
  `on` requires a persona file under `docs/agents/`).
- A repo that clones with `unknown` re-nudges every session until decided; `skip`
  persists the decision. This is intended.
- The persona-content work is deferred to the skill and a user-chosen template
  dir, so the binary carries no persona library and no dependency on one.

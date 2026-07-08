# ADR 0006: PI enforcement adapter

- **Status:** accepted
- **Date:** 2026-07-07

## Context

ADR 0002 modelled enforcement as an adapter registry and foresaw a PI adapter
"later"; until now `internal/enforce` held one concrete `Claude{}` and init/check
called it directly. PI (the Pi Agent Harness) differs from Claude: it reads
`AGENTS.md` natively once a project is trusted, and it has no settings-file
`SessionStart` hook — its injection point is a TypeScript extension subscribing to
the `session_start` lifecycle event.

## Decision

Extract an `enforce.Adapter` interface (`ID`/`PlanLine`/`Apply`/`Verify`) and add a
`PI{}` implementation that writes `.pi/extensions/gyroscope-context.ts`. On
`session_start` the extension injects the *non-hub* spokes (the hub is read
natively) via `pi.sendMessage(..., { deliverAs: "nextTurn" })`. A new `enforce`
config section selects adapters; Claude stays default-on, PI is opt-in. init and
check loop the enabled adapters. Claude's interface methods wrap its existing
merge/inspect logic — no change to the tested hook code.

gyroscope does not automate PI trust (`~/.pi/agent/trust.json`): that is a user
security decision, surfaced in the skill/README instead. PI is not added to the
doc-target registry, since it reads `AGENTS.md` directly and needs no pointer file.

## Consequences

- A new harness is now a drop-in `Adapter` plus a config flag; init/check need no
  per-harness branches beyond `enabledAdapters`.
- PI's extension is a managed file — regenerated and drift-checked from the enabled
  paths; editing it by hand is reported as drift.
- Enforcement is genuinely opt-in per harness (ADR 0002 realised), so repos that
  do not use PI never get a `.pi/` tree.

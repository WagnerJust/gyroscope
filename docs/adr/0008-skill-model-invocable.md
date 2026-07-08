# ADR 0008: The skill is model-invocable, gated by approval

- **Status:** accepted
- **Date:** 2026-07-08

## Context

The `/gyroscope` skill shipped with `disable-model-invocation: true` — a deliberate
choice (see the "user-invoked" design note) so the interview and the file writes it
drives only ever start from an explicit `/gyroscope`. The reasoning: a skill that
writes files and installs a hook should not fire on a stray mention.

That guarantee has a cost. A natural request — "get this repo up to date with
gyroscope", "adopt gyroscope here", "make `gyroscope check` pass" — did **not**
trigger the skill. The model fell back to driving the binary ad hoc: it usually
reached a reasonable result (the merge-safe `init` and the `check`/`--fix` loop make
that viable), but without the skill's structured interview, its reconciliation
discipline, or its guaranteed steps. The curated flow existed but was unreachable
from the phrasing users actually use.

Two ways to close the gap were considered:

1. **A second, model-invocable "sync" skill**, leaving `/gyroscope` user-invoked.
   Clean separation, but two skills to keep in sync and a split mental model for one
   product.
2. **Flip `/gyroscope` itself to model-invocable** and broaden its description to
   cover both setup and update.

The over-triggering risk that motivated the original flag is already contained by a
mechanism the skill has independent of invocation: the HARD-GATE. The skill must
present a plan and get approval before writing anything; `gyroscope check` and
dry-run `init` are read-only. So model-invocation changes *when the skill starts
thinking*, not *when it starts writing*.

## Decision

Remove `disable-model-invocation` from `skill/SKILL.md` and broaden its description
so a plain request ("get this repo up to date with gyroscope", adopt, migrate a
buckle-style hub, or fresh setup) triggers it. Keep — and strengthen — the HARD-GATE
so that being auto-invoked never bypasses the show-a-plan-then-get-approval step. Add
an explicit "Existing repo — reconcile, don't clobber" flow so the update path is a
guaranteed script, not improvisation. This reverses the earlier user-invoked
decision.

## Consequences

- The phrasing users actually use now reaches the curated flow: interview where
  needed, reconcile where a repo already has docs, converge to `check`-green.
- The write-safety guarantee no longer rests on invocation being manual; it rests on
  the HARD-GATE. That gate is now load-bearing and must survive edits — if it is ever
  weakened, the skill can write without consent. Tests/reviews should treat it as
  such.
- One skill still owns both setup and update, so there is no second artifact to drift
  — at the price that the skill's description must stay broad enough to trigger on
  update phrasing without being so broad it fires on unrelated mentions.
- `gyroscope check` becomes the contract an agent converges toward; the merge-safe
  `init --apply` (ADR 0007) and `check --fix` are what make an auto-invoked run safe
  by default.

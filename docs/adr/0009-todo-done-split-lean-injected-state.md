# ADR 0009: TODO/DONE split — lean injected state

- **Status:** accepted
- **Date:** 2026-07-09

## Context

The SessionStart hook `cat`s the whole state file into context every session, so
`TODO.md`'s size is a recurring per-session cost. As a repo's history accumulates,
the completed trail dominates the file, and the agent pays to re-read finished work
it will never act on again — the exact "setup/resumption is heavy" problem the state
spoke was meant to solve, inverted by growth.

A prior adoption (notwhoop) tried to fix this with an agent-invented split: a "lean
root `TODO.md`" plus a "detailed `docs/TODO.md`". That relied on discipline the
agents didn't keep — observed behavior was to treat the injected root file like a
normal todo and let it grow anyway, ending with two todo files and no smaller
injection. A convention with no mechanism doesn't hold.

The genuine alternatives were:

- **An in-file marker** (a `## Done` section inside `TODO.md`). Rejected: the hook
  `cat`s the *whole file*, so an in-file boundary still injects the archive. The cost
  is a function of file size, and one file has one size.
- **A separate archive file** read by no hook. The whole-file `cat` that makes an
  in-file marker useless is exactly what makes a second file work: excluding a path
  from the hook excludes its entire contents, mechanically, with no discipline
  required.

## Decision

Split the state spoke into two files by a mechanical status boundary:

- `TODO.md` = **open work only** (in flight / next). Injected by the SessionStart
  hook every session, so it must stay small.
- `DONE.md` = **completed archive**. Part of the standard — scaffolded by `init`,
  present in `standard.Plan`, hub-routed, and checked — but **never** added to
  `hookPathsFor`, so it is never injected.

Both ride under the existing `state` config toggle (`TODO.md` + `DONE.md` +
`.local/todo.md` are one unit); no new config key. The move rule — "when a task is
done, move its line to `DONE.md`" — is stated in the `TODO.md` header and the
`docs/agents.md` conventions, and given teeth by `gyroscope check`: past a threshold
of completed `[x]` items left in `TODO.md`, check emits an archive nudge. The nudge
is a **soft note, not drift** (it prints but exits 0) — an unarchived done item is
untidy, not a structural nonconformance, and housekeeping shouldn't break CI.

## Consequences

- The per-session injection stops growing with history: only open work is catted.
  The archive can grow without bound at no session cost.
- `DONE.md` is a new spoke genre: enforced + routed but not injected. The hook's
  cat-set and the hub's route-set are no longer the same set — the SessionStart term
  and the Spoke term in `CONTEXT.md` are updated to reflect it.
- The move rule is a convention, but now a checked one — the nudge converts drift-by-
  neglect into a visible signal without failing the build.
- Adoption of an existing repo must split its TODO state and consolidate any stray
  todo file into the root, so the standard's own reconciliation replaces the failed
  lean/detailed convention rather than inheriting its mess (`skill/SKILL.md`).

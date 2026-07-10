# ADR 0011: check --fix archives completed tasks — the nudge grows teeth

- **Status:** accepted
- **Date:** 2026-07-10

## Context

ADR 0009 split state into an injected `TODO.md` (open work) and a non-injected
`DONE.md` (history), and gave the "move a done task to `DONE.md`" rule teeth via a
`check` archive nudge — a soft note past a threshold of completed `[x]` items left in
`TODO.md`. In practice the teeth did not bite the actor. The nudge fires at `check`
time, to whoever *runs* check — usually a human — while the move is asked of the
mid-session agent that just finished a task and never runs check. The rule is stated
three times (the `TODO.md` header, the `docs/agents.md` conventions, the hub route)
and still needs manual reminding: a convention whose enforcement never reaches the
one who must act. That is the exact failure ADR 0009 diagnosed in the prior
lean/detailed split — *"a convention with no mechanism doesn't hold"* — reappearing
one level up. ADR 0009 mechanized the injection *cost* (a separate, un-injected file)
but left the *move itself* as pure convention.

The genuine alternatives for closing it:

- **A fourth prose statement** ("agents MUST move done items"). Rejected: the rule is
  already stated three times; more prose is more of the thing already failing.
- **A `Stop` hook that nags each turn.** Rejected: it is the session-start nag on
  repeat — same failure class (a reminder competing for attention), more friction, and
  quickly tuned out.
- **A SessionStart hook that auto-moves `[x]` lines.** Rejected as the default: it
  turns a deliberately *read-only* hook into one that mutates git-tracked files at
  session start (surprise diffs before the user has acted), and the move carries light
  judgment — `DONE.md` groups items under headings — that a blind append flattens.
- **A standalone `gyroscope archive` subcommand.** Rejected as the primary path: a
  separate command is a separate thing to remember, reintroducing the very gap. (The
  transform is still exposed as a reusable package for a future subcommand if wanted.)
- **Fold the move into `check --fix`.** Chosen — see below.

## Decision

Make the archive move the convergence half of the nudge: `check --fix` moves completed
top-level tasks out of `TODO.md` and into `DONE.md`. The nudge was the only `check`
finding `--fix` could not fix; now it can. Archival rides the converge-to-green loop
the agent already runs, so there is no new rule to remember.

- **Mechanics live in the binary; judgment stays with the author.** A pure
  `internal/archive` package does the transform: a *top-level* `- [x]` line (plus its
  indented sub-tree) is a completed task and moves; a nested `[x]` under an unfinished
  parent stays. Moved blocks are inserted at the top of `DONE.md`'s `## Completed`
  section (newest on top, per the `DONE.md` convention) — no dates, keeping it
  deterministic and CI-safe, since git already records when. The residual judgment
  (curating `DONE.md` groupings) degrades gracefully to a re-curatable list.
- **A second deliberate non-`WriteGuarded` write.** The move rewrites two files that
  already exist, so it uses a new `fsutil.WriteAtomic` (temp file + rename, overwrite)
  rather than `WriteGuarded`'s O_EXCL clobber-refusal — the same reasoning as the
  persona mirror (ADR 0010), now centralized in `fsutil`. It is loss-safe: `DONE.md`
  is written before `TODO.md`, so a failed second write duplicates items (still in
  both) rather than losing completed work, and a re-run converges.
- **Gated and quiet.** It runs only under `--fix` and only when the `state` spoke is
  on, and is a no-op (nothing written) when no top-level `[x]` item is present.

## Consequences

- The nudge stops being advice-only: `check --fix` both detects and clears the backlog,
  so after a fix the injected `TODO.md` is lean without anyone hand-editing two files.
- gyroscope now rewrites *user-owned* content files (`TODO.md`/`DONE.md`), not just its
  own scaffolds and mirror — a widening of scope. It is bounded: only the `--fix` path,
  only the `state` spoke, only a lossless line move, atomic and DONE-first.
- `fsutil.WriteAtomic` is the shared home for gyroscope's deliberate overwrite writes;
  the persona mirror's local atomic writer (ADR 0010) can later fold into it.
- Enforcement still requires *running* `check --fix` — not a hard guarantee — but that
  is already the documented end-of-task step, a far smaller ask than a manual two-file
  move. If a hard gate is ever wanted, the lever is flipping the nudge to drift (fail
  CI past a threshold); ADR 0009 chose soft deliberately, and a CI gate would still
  punish the human, not the actor, so mechanization is preferred to gating.

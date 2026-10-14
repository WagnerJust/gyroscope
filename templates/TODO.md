# TODO

> Living, hand-maintained state file — the repo-wide source of truth for **open
> work only** (what's in flight and what's next). Tracked in git, and injected by
> the SessionStart hook so a fresh agent chat resumes where the last one stopped
> instead of re-deriving context.
>
> **One line per open item.** Narrative — what was tried, why, what it turned out
> to mean — belongs in a `DONE.md` entry or an ADR, with a pointer here if needed.
> **When a task is done, move its line to `DONE.md`** — or let `gyroscope check
> --fix` do it: it archives completed top-level `[x]` items into `DONE.md` for you.
> The hook injects this file only while it fits a 9,000-byte session budget; past
> that it is skipped with a one-line pointer and `gyroscope check` reports drift.
> The completed archive lives in `DONE.md` (routed, not injected). Keep it
> current; keep it terse.

**Legend:** `[ ]` not started · `[~]` in progress · `[x]` done (then move to `DONE.md`)

## Now
- [ ] Replace with what's actively in flight.

## Next
- [ ] Replace with what's queued after Now.

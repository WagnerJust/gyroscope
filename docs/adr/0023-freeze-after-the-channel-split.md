# ADR 0023: Freeze the standard and the binary after the channel split

- **Status:** accepted
- **Date:** 2026-10-15

## Context

On 2026-10-14 the tool was assessed against the two research surveys in
`~/Side/research` (`llms-and-agents/`, `agentic-coding/`), prony's committed
findings, and the session transcripts of every adopting repo. The question was
whether the evidence supports what the binary adds over a hand-written
conventions file plus `TODO.md`/`DONE.md`. It does not:

- The only controlled study of instruction files (Gloaguen et al., ETH, arXiv
  2602.11988, 138 tasks) finds context files do not generally improve task
  success, add about 20% cost, and that repository overviews do not help; the
  useful content is short corrective rules for non-standard practices. Nothing in
  either survey tests routing, state injection, personas, ADRs or drift tooling.
- The hub-and-spoke premise fails at both ends in this repo's own transcripts.
  prony measured that an un-injected hub is not read (0 of 8). Where it is
  injected, the spokes it routes to are opened in under 10% of sessions
  (docs/agents.md at most 8 of 92 prony sessions and 12 of 120 carsearchhub
  sessions; CONTEXT.md 9 and 6; CONTRIBUTING.md 1 and 2).
- State injection was not delivering before ADR 0022 (hook output truncated in
  every active repo) and the state it carries is not what sessions use to resume:
  2 of 50 hand-typed carsearchhub openers mention open work; resumption happens
  through `/next`, which outlines `TODO.md` itself.
- The cross-harness adapters (`enforce.pi/opencode/codex/cursor`) are off in all
  ten adopting repos. Eight of twenty-two ADRs and the two largest `DONE.md`
  sections are that work.
- `check` reported every repo conformant through three months of truncation,
  because it compared the configured hook string and never its output. That is
  fixed in ADR 0022, but it is the one measurement the tool made, and it was
  wrong for the entire period.
- Build cost from transcripts: twelve build sessions, 1.86M output tokens, 5.4%
  of all output across Side projects, 143 commits of which 123 are from July 2026.

What the evidence does support is small: a short, corrective `docs/agents.md`,
`TODO.md`/`DONE.md` as the human's backlog, personas invoked by slash command
(the one artifact that is used: carsearchhub 47 of 120 sessions), and
programmatic guards where a rule can be a predicate. None of that needs the
binary to exist, but the binary already exists and converging ten repos on one
layout is cheap to keep and costly to unwind.

The genuine alternatives were to keep building (path-scoped rules, `## Now`
extraction, terse output, the deferred items in `TODO.md`) or to delete the tool
and the standard. The first spends effort on mechanisms with no evidence behind
them; the second throws away a working layout and a `check` that now measures
delivery, for no measured gain.

## Decision

Freeze. The standard and the binary stay as they are at ADR 0022:

- **No new spokes, adapters, checks or toggles.** The `Later` section of
  `TODO.md` is closed, not deferred. A new item needs a measured failure in a
  transcript or an outcome-level study, not a design argument.
- **`TODO.md` is the human's backlog, not agent state.** It is not trimmed to
  fit the hook budget. The budgeted hook injects it when it fits and says so when
  it does not; `check` reporting it over budget is information, not a defect.
- **Bugs that lose data or install to the wrong place are still fixed** (the
  `install-skill` config-dir fix landed with this ADR). Nothing else is.
- **Re-measure routing on or after 2026-11-15** from the transcripts under
  `~/.claude-personal/projects/`. If spoke opens stay under 10% of sessions in
  repos where the hub is imported in full, fold the routed spokes into
  `docs/agents.md` or delete them, and record that in a new ADR.
- **Guards live outside gyroscope.** Branch, push, unattended and test gates are
  Claude Code hooks and slash-command steps maintained in `~/.claude-personal`
  and `~/Side/setup`; gyroscope does not grow an enforcement layer (ADR 0004
  already put CI gates out of scope).

## Consequences

Easier: adopting repos stop accumulating structure; the ten converged repos
stay converged with `check --fix` and nothing more. The transcript measurements
that produced this decision are kept as scripts in `claude-tokens`, so the
re-measure is a command, not a project.

Harder: carsearchhub's `docs/agents.md` (1,258 lines) has to be cut by hand to
the corrective subset; the path-scoped-rules phase that would have carried the
rest is closed. Any harness other than Claude Code gets the hub only through
the pointer files, which the measurements say are not followed.

Committed to: not reopening this without a measurement. The question the
freeze cannot answer, because nothing measured it, is whether a consistent
layout across repos has a value of its own for one maintainer. If it does, the
freeze costs nothing; if it does not, the standard is a habit, and the
re-measure date is when that gets asked again.

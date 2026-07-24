# ADR 0013: ADR enforcement — encourage authorship, police structure as drift

- **Status:** accepted
- **Date:** 2026-07-23

## Context

gyroscope scaffolds `docs/adr/` (TEMPLATE + 0001) and routes to it from the hub, but
never encouraged *writing* ADRs nor policed the ones written. Two gaps followed from
that. First, nothing in the injected hub told an agent to record a significant
decision, so the discipline depended on a route an agent might never follow.
Second — and concretely — a shipped ADR once referenced a sibling ADR number that
did not exist (a dead link); no check caught it because no check read the ADR set at
all. Numbering could drift (gaps, duplicates) and a record could ship missing its
`Status` or `Date` with the same silence.

The forces in tension:

- **Encourage vs over-produce.** The TEMPLATE already guards against ADR sprawl
  ("write one only when the decision is hard to reverse, surprising, and a real
  trade-off"). A usage directive must echo that guard, not push agents to file an
  ADR for every change.
- **How hard to police, and where the finding lands.** The genuinely open question
  (flagged TBD when this was filed): are structural findings (numbering gap, missing
  `Status`/`Date`, dangling reference) *soft notes* — like the archive nudge (ADR
  0011), which never affects the exit code — or *drift* that fails CI?
- **What is checkable.** "Did you write enough ADRs" is unknowable and must never be
  checked. Numbering contiguity, field presence, and referent existence are
  objective and checkable.

## Decision

Add both halves, mirroring the personas/attribution pattern (a hub directive plus a
`check`), gated on the existing `adr` spoke.

- **Usage (prompt-injection).** A standing managed-hub directive, `## Architecture
  decisions`, spliced at a `<!-- gyroscope:adr-directive -->` marker and rendered
  only when the `adr` spoke is on. It instructs agents to record a significant
  decision as an ADR and reference it from the code it shapes, and it repeats the
  TEMPLATE's own three-part guard so it encourages without inflating the record
  count. `standard.ADRDirective` is the single source of truth, shared by the
  renderer and by `check` (which verifies the region still contains it).
- **Upkeep (structural check).** A pure `internal/adr` package checks the set:
  contiguous numbering from 0001, a `**Status:**` and `**Date:**` field on every
  numbered record, and no reference (`ADR-NNNN` / `ADR NNNN`) to a number with no
  file. TEMPLATE.md is excluded from the numbered set and the reference scan, so its
  `ADR-XXXX` placeholders never register as dead links.
- **Findings are drift, not notes.** This resolves the TBD. A numbering gap, a
  missing field, or a dead reference is an objective defect analogous to the existing
  dangling-route check — and catching the real dead-link bug in CI was the whole
  point — so these fail `check` (exit `exitDrift`), unlike the archive nudge's soft
  note. The *directive* stays soft (it encourages judgment); the *structure* is
  policed hard.

## Consequences

- `check --fix` cannot author these findings (it can't safely renumber, date, or
  invent a referent), so they surface as persistent drift until a human resolves
  them — the same shape as an unfilled `{{...}}` placeholder. This is deliberate:
  authorship is judgment, and `--fix` never fabricates content.
- The "soft" framing in the original task refers to the directive and to *not*
  counting ADRs — not to the structural checks, which are hard by design.
- The reference scan is intentionally conservative (uppercase `ADR`, a `\d` guard) to
  keep false positives near zero; a lowercase or unusual reference form may slip past,
  costing a missed finding rather than a wrong CI failure.

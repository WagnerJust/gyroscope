# ADR 0017: code review is an opt-in persona preset, skill-scaffolded

- **Status:** accepted
- **Date:** 2026-07-24

## Context

Setting up a repo's agent personas from scratch (ADR 0005/0010) is an open-ended
interview. But one persona configuration recurs across repos: a curated,
multi-persona **code-review** suite — layer-aligned reviewer specialists, a
synthesizer that aggregates without pre-filtering, and an intent-verifier that
grounds every finding against the repo's decision record. Rebuilding it by hand each
time is repeated, error-prone setup. It is the persona mechanism used as a canonical,
opinionated preset rather than a blank interview.

Two design questions had to be settled:

1. **Where does it live — binary or skill?** The binary never authors persona
   content (the binary/skill split, ADR 0003): it writes structure and re-encodes
   its own mirror files, but reviewer personas are opinionated *content*, customized
   per repo. So the binary cannot own the bundle.
2. **A new spoke/toggle, or a preset inside the persona flow?** The bundle spans
   genres — several personas, a synthesizer, an intent-verifier, an orchestrator
   runbook, and a grounding ADR — broader than one spoke. Modeling each as a config
   spoke would multiply toggles for one feature.

## Decision

Code review is an **opt-in preset that redirects the persona flow**, driven by a
single `gyroscope.json` flag `codeReview` (default false). The binary only records
the flag; the `/gyroscope` skill authors the bundle when it is set (or when the user
asks for a review suite), instead of the from-scratch persona interview.

The skill scaffolds, **repo-tuned** (only the specialists that apply to this repo):

- Layer-aligned `*-reviewer.md` personas, each with a `description:` trigger and an
  `appliesTo:` glob; `security-reviewer` always runs.
- `synthesizer.md` — aggregates, never pre-filters.
- `intent-verifier.md` — runs last, grounding each finding against `docs/adr/`,
  `TODO.md`, `DONE.md`, and the spokes (UPHELD / OVERRULED-with-citation /
  NEEDS-DECISION), and scanning for intent-violations the reviewers missed. This is
  the load-bearing piece: it catches the "technically right but wrong for this repo"
  false positive and the silent regression against a prior decision. It is distinct
  from `reality-checker` (which runs checks to prove a claim).
- A `code-review.md` orchestrator runbook.
- A grounding ADR recorded in the *adopting* repo, so the intent-verifier's rationale
  is on the record.

The reviewer personas are ordinary personas: they register into every enabled
harness registry (ADR 0010/0014/0015) and are policed by the existing persona and
ADR-hygiene checks — no new enforcement surface is added for the preset itself.

## Consequences

- The recurring review-suite setup collapses to one flag plus the skill's canonical
  definition, instead of a from-scratch rebuild per repo.
- The canonical menu comes from a web app, so most repos use a subset (a Go CLI:
  backend + security + cicd). The skill customizes and prunes per repo — the preset
  is a starting point, not a fixed set imposed on every repo.
- It complements, not replaces, a generic review runner (e.g. Claude's built-in
  `/code-review`): this supplies the repo-specific reviewers and the intent-grounding
  a generic runner lacks.
- The binary stays content-free: adding `codeReview` is one config field, no
  templates, no new writer — the split holds.

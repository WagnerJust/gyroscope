# ADR 0001: Record architecture decisions

- **Status:** accepted
- **Date:** 2026-07-04

## Context

This repo uses agents heavily. Agents (and humans) repeatedly re-litigate choices
whose rationale was never written down, and reverse decisions without knowing why
they were made.

## Decision

We record significant architecture decisions as ADRs in `docs/adr/`, using
`TEMPLATE.md`. We write one only when the decision is hard to reverse, surprising
without context, and the result of a real trade-off.

## Consequences

- Future agents read `docs/adr/` (routed from `AGENTS.md`) before changing shaped
  code, instead of guessing.
- We accept the small overhead of writing ADRs for the decisions that warrant it —
  and deliberately skip the ones that don't.

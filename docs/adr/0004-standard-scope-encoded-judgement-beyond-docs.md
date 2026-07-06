# ADR 0004: The standard spans encoded judgement, not just docs

- **Status:** accepted
- **Date:** 2026-07-06

## Context

gyroscope's original standard was a set of **agent docs**: the `AGENTS.md` hub,
prose spokes it routes to (`CONTEXT.md`, `docs/agents.md`, ADRs, personas), the
pointer files, and the `SessionStart` hook. "Spoke" meant "a markdown doc the hub
links to." Two pressures pushed the standard past that definition:

1. **ACMM Level 2 completeness.** plumbline scores a repo against the AI Codebase
   Maturity Model. L2 ("Instructed") is not just instruction prose — it also
   counts a **PR-template checklist**, **commit-message rules**, and a
   **contributor guide**. A filled gyroscope repo nailed the instructions signal
   but shipped none of the process artifacts, so it scored L1, not L2.
2. **The "issues with agentic coding" problem doc.** Its Issue #2 is that setup
   and — critically — *resuming existing work in a fresh chat* is heavy. Docs
   orient an agent to the rules; nothing oriented it to *where the work stands*.

The genuine alternative was to keep the standard doc-only and let other tools or
manual steps supply process files and state. That keeps the vocabulary tidy
("the standard = agent docs") but leaves gyroscope short of a clean L2 and blind
to Issue #2's resumption half — the exact problems it exists to solve.

## Decision

Expand the standard to include **encoded judgement in whatever form the tooling
enforces it**, not only hub-routed prose. Concretely, the standard now also
writes:

- **Process artifacts.** `CONTRIBUTING.md` (a hub-routed doc spoke, scoped to
  process and deferring to `docs/agents.md` for conventions so the two don't
  drift); `.github/pull_request_template.md` and `.gitmessage` (enforcement
  genre — Git/GitHub apply them at commit/PR time, so they are written but **not**
  hub-routed).
- **State files.** A tracked, repo-wide `TODO.md` (hub-routed) and a gitignored,
  personal `.local/todo.md`, both injected by the `SessionStart` hook so a fresh
  session resumes from current progress.

The unifying principle: the standard is *durable encoded judgement plus the
mechanisms that make an agent encounter it*, across three enforcement surfaces —
the hub + session hook, Git/GitHub, and hook-injected state — not merely markdown
the hub links to. Each artifact is classified by how it reaches the agent, which
determines whether it carries a hub route.

## Consequences

- A filled gyroscope repo now scores plumbline **L2 = 1.0** (verified), and the
  standard addresses Issue #1 (program docs into the tools) and Issue #2b (resume
  from injected state).
- The "spoke" vocabulary generalizes: config toggles in `SpokeSet` now gate
  non-routed artifacts (`prTemplate`, `commitConvention`) and state files
  (`state`), not just routed docs. `CONTEXT.md`'s glossary should be read with
  that widened sense.
- The scope boundary is now explicit and must hold: **L3+ of the ACMM** (CI
  gates, coverage, feedback loops) stays out — it needs runtime/CI infra
  gyroscope deliberately doesn't touch and belongs to plumbline. **Output shaping**
  (Issue #3, wordy explanations → rubber-stamping; e.g. `caveman`) also stays out
  — gyroscope shapes what an agent *reads*, not what it *says*.
- More surface to keep coherent: three enforcement channels instead of one. The
  binary/skill split is unchanged — the binary writes structure verbatim; the
  skill fills content.

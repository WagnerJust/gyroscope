# ADR 0012: gyroscope-maintained repos are first-class for devs without gyroscope

- **Status:** accepted
- **Date:** 2026-07-10

## Context

Adopting gyroscope should not tax contributors who never installed it. At runtime a
gyroscope repo is already tool-free: the SessionStart hook is pure `cat … 2>/dev/null
|| true` (no binary), and the hub, spokes, pointer files, and `.claude/agents/` mirror
are plain Markdown any agent or human reads natively. The gap is **maintenance +
legibility**, not usage:

- A stranger opening the repo sees `gyroscope.json`, `<!-- gyroscope:managed -->`
  markers, `.local/`, and hub text mentioning `/gyroscope` — all meaningless without
  the tool, with nothing explaining that the tool is optional.
- If they hand-edit a managed region or add a spoke, nothing tells them it is safe
  (that a maintainer's or CI's next `check --fix` reconciles it), so they either avoid
  the docs or fear breaking them.

The alternatives for delivering the explanation:

- **A write-once section in the `CONTRIBUTING.md` template.** Rejected as primary: only
  *new* adopters get it, and a dev can silently drift or delete it — not guaranteed for
  the already-maintained repos this is meant to serve.
- **A managed section in the hub (`AGENTS.md`).** Rejected as the home: the hub is
  catted into every session, so a full human-facing explainer there taxes every agent
  run — the lean-injected-hub principle (ADR 0009). A single pointer line is affordable;
  the full note is not.
- **A managed block in `CONTRIBUTING.md`.** Chosen: human-facing (the conventional home
  a contributor looks), *not* injected (no per-session cost), and gyroscope-owned so it
  is guaranteed and self-converging into existing repos.

Delivering it required a small generalization: managed regions had been a hub-only
mechanism.

## Decision

Serve both audiences, and generalize the managed region so a spoke can carry one:

- **Managed regions are no longer hub-only.** Any planned file whose standard content
  carries `<!-- gyroscope:managed -->` … `<!-- /gyroscope -->` is merged in place, not
  treated as an all-or-nothing collision. The three hub-hardcoded spots were generalized:
  `standard.InjectManaged` takes a `dest`; converge's classifier keys the MERGE path on
  "want has a managed region" instead of `dest == "AGENTS.md"`; and `check` verifies each
  non-hub managed spoke's on-disk region byte-for-byte against the standard. The hub keeps
  its own semantic checks (routes + personas directive) because its region is
  config-rendered; a spoke's region is static, so byte-equality is the right test.
- **The contributor note lives in `CONTRIBUTING.md`** as that managed block: these docs
  are gyroscope-maintained; you do not need gyroscope to read them or contribute; what
  the markers / `gyroscope.json` / `.local/` mean; hand-edits inside a region reconcile
  on the next `check --fix`; and an optional zero-install escape hatch — `go run
  github.com/WagnerJust/gyroscope/cmd/gyroscope@latest check .` (gyroscope is
  dependency-light and `go install`-clean, so no prior install is needed).
- **The hub carries a one-line pointer** inside its managed region ("Not using
  gyroscope? … see `CONTRIBUTING.md`"), so agents reading the always-injected hub learn
  the tool is optional at negligible token cost.

For an existing adopter whose `CONTRIBUTING.md` predates the block, `MergeManaged`
appends the wrapped region at EOF, preserving all their prose above it; `check --fix`
performs it. New repos get the full template including the block.

## Consequences

- Maintained repos become legible to and contributable-by devs who never install
  gyroscope, and the promise is guaranteed (checked + converging), not just documented
  for new adopters.
- The managed-region mechanism is now a general spoke feature, not a hub special case —
  reusable for future gyroscope-owned blocks in other spokes. `check` gains a byte-equal
  verification for non-hub managed spokes; the hub stays semantically checked.
- A second file (`CONTRIBUTING.md`) can now drift on its managed region and be
  reconciled — same detect-and-`--fix` contract as the hub, extended.
- The zero-install `go run …@latest` path is documented but unverified in CI (it needs a
  published module/tag); it is presented as optional, and a maintainer or CI keeps the
  repo conformant regardless.

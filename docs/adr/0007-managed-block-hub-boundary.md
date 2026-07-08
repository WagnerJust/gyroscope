# ADR 0007: The hub has a managed-block boundary

- **Status:** accepted (breaking change to the hub format)
- **Date:** 2026-07-08

## Context

gyroscope wrote the whole of `AGENTS.md` and treated it as gyroscope's to own:
`init` refused to touch a pre-existing hub without `--force`, and `--force` then
clobbered whatever the user had there. That is the wrong shape for the two most
common real cases:

- **Adoption on a repo that already has an `AGENTS.md`** (e.g. a buckle hub, or a
  hand-written one). The user is forced to choose between refusing (gyroscope does
  nothing) and clobbering (gyroscope destroys their curated content).
- **A user who wants to add their own prose to the hub** — a project preamble, a
  "read this first" note. Any such addition made re-`init` refuse, and made
  `check` see the extra content as drift.

ADR 0003 already carved out a bounded exception to "no rendering engine": the
binary assembles the `## Routes` list from config at a single
`<!-- gyroscope:routes -->` marker. That marker proved the idea — gyroscope can
own a *delimited* slice of the hub rather than the whole file — but it was scoped
to one list. The genuine alternatives were: (1) keep whole-file ownership and lean
on `--force` (status quo — rejected, above); (2) attempt a 3-way content merge of
the user's hub against gyroscope's (large, and out of scope by decision — see
Consequences); (3) generalize the marker into a full **managed region** the binary
owns and re-writes, leaving everything outside it to the user.

## Decision

The hub gains a **managed region**: the content between `<!-- gyroscope:managed -->`
and `<!-- /gyroscope -->`. gyroscope owns only that region — the routes, the
pointer-files list, and the personas directive live inside it. Everything outside
the markers is the user's: gyroscope never writes it, and `check` never reads it.

- The template wraps gyroscope's body in the two markers. A fresh `init` writes
  the whole hub (heading plus managed block), exactly as before but delimited.
- `standard.MergeManaged` swaps one file's managed region for another's, touching
  nothing outside the markers. This is the in-place merge path the classifier's
  `MERGE` state (ADR — D1) and the merge-safe apply (D3) use to bring an existing
  hub's region current without disturbing surrounding prose.
- `check` extracts the managed region first and runs the route/personas
  comparison against *only* that region. A hub with no well-formed managed region
  is reported as drift with a migrate hint; user content outside the region — and
  any route-like bullet in it — is invisible to `check`.

This is a **breaking change to the standard's hub format**: a hub written by an
older gyroscope (no markers) now reads as `MERGE` (from `init`) or as
"managed region not found" drift (from `check`) until re-`init` migrates it.

## Consequences

- Adoption stops being all-or-nothing: gyroscope drops into a repo that already
  has an `AGENTS.md` by injecting its managed block and preserving the user's
  content. The buckle-migration pain, the custom-route-trips-`check` edge, and the
  re-`init` collision refuse all dissolve into one boundary.
- The "no rendering engine" line holds: the managed region is still assembled from
  fixed strings gated by config (ADR 0003), now delimited rather than whole-file.
- We are **not** doing a 3-way content merge. The managed block gets ~90% of the
  value (own a delimited slice, leave the rest alone) for ~10% of the code; a real
  content merge stays deferred unless blocks prove too weak (see TODO "Out of
  scope").
- The markers are load-bearing: hand-deleting or reordering them makes gyroscope
  lose its region. They are HTML comments, so they are invisible in rendered
  Markdown, but editing the hub around them must keep them intact — `check`
  enforces this.
- gyroscope's own `AGENTS.md` is migrated to the managed-block form in this change
  so the repo stays conformant on itself.

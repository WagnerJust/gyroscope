# ADR 0003: The hub prunes routes for disabled spokes

- **Status:** accepted (supersedes the static-hub-with-hedge approach)
- **Date:** 2026-07-06

## Context

The hub (`AGENTS.md`) routes agents to spokes. Spokes are toggleable in
`gyroscope.json`, but the hub template was **static** — it listed all routes
regardless of which spokes were enabled. Disabling a spoke skipped writing its
file but left its route in the hub, a dead link. The original mitigation (post-MVP
review #1) was an honest hedge at the top of `## Routes`: *"Spokes are optional:
if one is disabled, its route points to a file that wasn't generated."*

That hedge is a poor fit for the audience. The whole point of gyroscope is to
stop agents from having to guess: an agent reads the hub, follows a route, and
finds the spoke. A dangling route makes the agent open the hub, try to read a
spoke, and discover it isn't there — exactly the wasted, confidence-eroding step
gyroscope exists to remove. A hedge asks the *reader* to compensate for a defect
the *writer* could just avoid.

The genuine tension is with a load-bearing principle: **the binary writes
scaffolds verbatim and does NOT template — there is no rendering engine**
(`docs/agents.md` §"Do NOT"). Pruning routes means the binary must assemble the
route list from config rather than copy a fixed block. The alternatives were:
(1) keep the static hub + hedge (status quo — rejected, above); (2) have the
companion skill prune the hub after the binary writes it (splits one concern
across the binary/skill boundary and makes a non-interactive `init` produce a
wrong hub); (3) let the binary render the enabled routes.

## Decision

The binary renders the hub's `## Routes` list from config. `templates/AGENTS.md`
carries a single `<!-- gyroscope:routes -->` marker; `renderRoutes` replaces it
with one bullet per **enabled** built-in spoke (in canonical order) followed by
one per custom spoke. Disabled spokes are simply absent, so the hub never points
at a file that wasn't written. The "spokes are optional" hedge is removed.

This is deliberately **not** a templating engine. It is a contained assembly of
fixed route strings gated by booleans — the same shape as the `entries` table in
`Plan()` that already gates which files get written, and a generalization of the
pre-existing custom-routes marker (which this replaces). No variable
substitution, no per-repo content rendering; the route prose is fixed in the
binary. The binary/skill split is intact: routes are *structure* (the binary's
job), not interviewed *content* (the skill's job).

## Consequences

- The hub is always internally consistent with `gyroscope.json`: every route
  points at a spoke that exists. No dead links, no reader-side hedge.
- The route↔toggle mapping lives in one place (`renderRoutes`), next to the
  file↔toggle mapping in `Plan()`. Adding a spoke means touching both, together.
- The "no rendering engine" rule now has an explicit, bounded exception: fixed
  strings selected by config. Future changes must stay on this side of the line —
  selecting among fixed structural strings is fine; substituting repo-specific
  content in the binary is not (that remains the skill's job).
- Route prose now lives in Go, not the `AGENTS.md` template. Editing a route's
  wording is a code change, guarded by tests, rather than a template edit.
- Per-file "may not exist" notes stay where they describe a *different* fact: a
  gitignored `.local/` spoke can be enabled yet absent from a fresh clone. That
  is not a dangling route, so those hedges remain.

# gyroscope — build TODO

> Living build tracker, hand-maintained (not via the todo tool). **Open work
> only** — this file is injected by the SessionStart hook every session, so keep
> it small. When a task is done, move its line to `DONE.md` (the routed, not-
> injected archive). Design docs & scratch live in `.local/` (gitignored).

**Legend:** `[ ]` not started · `[~] `in progress · `[x]` done (then move to `DONE.md`)

## Next
- _(empty — no open work)_

## Known non-issues (kept for context; not action items)
- [ ] (note, not fixed) `go build ./cmd/gyroscope` reports `dev (commit none, built unknown)` — no ldflags; `make build` is the versioned path. Expected, documented as the "quick local binary."

## Later — deferred (explicitly out of MVP)
- [ ] **PI persona sub-agents (extension-built) — parity for the `.claude/agents/` mirror.** PI has NO native sub-agent registry (by design — its README: "No sub-agents… build your own with extensions"), so there is nothing to mirror `docs/agents/` into the way F1–F4 do for Claude. Sub-agents are *buildable* via PI's extension API, though. To give PI parity, extend the gyroscope PI extension (`.pi/extensions/gyroscope-context.ts`) to register a sub-agent tool that, per `docs/agents/` persona, spawns `pi -p <prompt> --tools <persona tools>` with the persona as the system prompt. This is BUILDING runtime behavior (a harness feature), not a byte-copy mirror — larger scope than the Claude mirror, opt-in, gated on `enforce.pi` + `personas == on`. Keep `docs/agents/` canonical; the extension reads it live. See ADR 0010 (Claude mirror) for the docs-vs-registered rationale this would extend to PI.
- [ ] **Zed enforcement adapter** — active `enforce.Adapter` for Zed (the passive `.rules` doc-target pointer already exists; this is the force-inject side, parallel to Claude's SessionStart hook / PI's `session_start` extension). Investigate Zed's injection mechanism (does its agent support a session-start hook / rule that force-reads the hub, or is `.rules` native-read only?). If native-read only, there may be nothing to enforce — document that outcome. Wire behind the `enforce` config section (`zed`, opt-in) if a mechanism exists.
- [ ] **Cursor enforcement adapter** — active `enforce.Adapter` for Cursor (passive `.cursorrules` pointer already exists). Investigate Cursor's mechanism: `.cursor/rules/*.mdc` with `alwaysApply: true` (always-injected project rules) and/or Cursor hooks. An always-applied rule that force-reads the hub would be the enforcement analog. Wire behind the `enforce` config section (`cursor`, opt-in).
- [ ] **Terse output for gyroscope repos (toggleable) — deferred; caveman manually for now.** Brainstormed 2026-07-10; full option analysis in `.local/design/2026-07-10-terse-output-options.md`. Shortlist: recommend **Option 5, adapter-shaped hybrid** — canonical `docs/terse.md` (hub-routed) that each enforce adapter registers natively, mirroring the persona pattern (ADR 0010): Claude adapter writes `.claude/output-styles/terse.md` (native output-style; `keep-coding-instructions: true` **required** or it strips coding discipline) + `outputStyle` in the settings.json it already merges; PI adapter injects the doc; pointer harnesses doc-routed. Gated on a `terse` toggle in `gyroscope.json` (default off; on for gyroscope = dogfood). Content: original wording informed by caveman (MIT — reuse allowed with attribution, but plugin-shaped; keep code/commands/**error strings** byte-exact, never terse-ify security/irreversible/ambiguous). Open decisions in the design note: scope 5a (Claude-first) vs 5b (full); toggle semantics — (A) sticky opt-out via output-style vs (B) ephemeral per-session, which needs a launch-gated cat-hook instead (output-styles can't do ephemeral). Core tradeoff: output-style = strong persistence + zero added input tokens but sticky-only toggle; cat-hook = weaker persistence + input cost but true per-session gating.

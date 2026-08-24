# gyroscope — build TODO

> Living build tracker, hand-maintained (not via the todo tool). **Open work
> only** — this file is injected by the SessionStart hook every session, so keep
> it small. When a task is done, move its line to `DONE.md` (the routed, not-
> injected archive). Design docs & scratch live in `.local/` (gitignored).

**Legend:** `[ ]` not started · `[~] `in progress · `[x]` done (then move to `DONE.md`)

## Next
- (all build-tracker "Next" items complete — see `DONE.md`. Cross-harness artifact sync is done across all 6 phases and live-validated on opencode + pi.)

## Known non-issues (kept for context; not action items)
- [ ] (note, not fixed) `go build ./cmd/gyroscope` reports `dev (commit none, built unknown)` — no ldflags; `make build` is the versioned path. Expected, documented as the "quick local binary."

## Later — deferred (explicitly out of MVP)
- [ ] **Residual persona tool-fidelity gaps (after the tool-map follow-up — ADR 0015 update 2026-07-24).** The `tools:` allowlist now carries through where each harness supports it (OpenCode `permission` block; PI `--tools` with verified names; Codex `sandbox_mode` read-only heuristic). Residual, all low priority: (1) **web tools** — WebFetch/WebSearch have no PI built-in (dropped) and are cleanly mapped for OpenCode but not PI; (2) **Codex is coarse** — sandbox mode only, no per-tool allowlist, and only tightens (read-only), never restricts a workspace-write persona further; (3) **verify-on-target** — Codex `sandbox_mode` honored-in-agent-TOML and PI's exact bad-`--tools` behavior are UNCONFIRMED from docs alone (issue openai/codex#15250); smoke-test against the real `codex`/`pi` versions. Also: Codex `commit_attribution` is a young feature — pin against the target `codex` version before relying on it.
- [ ] **Terse output for gyroscope repos (toggleable) — deferred; caveman manually for now.** Brainstormed 2026-07-10; full option analysis in `.local/design/2026-07-10-terse-output-options.md`. Shortlist: recommend **Option 5, adapter-shaped hybrid** — canonical `docs/terse.md` (hub-routed) that each enforce adapter registers natively, mirroring the persona pattern (ADR 0010): Claude adapter writes `.claude/output-styles/terse.md` (native output-style; `keep-coding-instructions: true` **required** or it strips coding discipline) + `outputStyle` in the settings.json it already merges; PI adapter injects the doc; pointer harnesses doc-routed. Gated on a `terse` toggle in `gyroscope.json` (default off; on for gyroscope = dogfood). Content: original wording informed by caveman (MIT — reuse allowed with attribution, but plugin-shaped; keep code/commands/**error strings** byte-exact, never terse-ify security/irreversible/ambiguous). Open decisions in the design note: scope 5a (Claude-first) vs 5b (full); toggle semantics — (A) sticky opt-out via output-style vs (B) ephemeral per-session, which needs a launch-gated cat-hook instead (output-styles can't do ephemeral). Core tradeoff: output-style = strong persistence + zero added input tokens but sticky-only toggle; cat-hook = weaker persistence + input cost but true per-session gating.

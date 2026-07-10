# gyroscope — build TODO

> Living build tracker, hand-maintained (not via the todo tool). **Open work
> only** — this file is injected by the SessionStart hook every session, so keep
> it small. When a task is done, move its line to `DONE.md` (the routed, not-
> injected archive). Design docs & scratch live in `.local/` (gitignored).

**Legend:** `[ ]` not started · `[~] `in progress · `[x]` done (then move to `DONE.md`)

## TODO/DONE split — lean injected state (2026-07-09)
> Problem (notwhoop dogfood): the SessionStart hook cats the WHOLE state file every
> session, so a large TODO.md = large injection cost every session — and the agent-
> invented "lean root + detailed docs" split relied on discipline agents ignore
> (observed: they use the injected root one like normal). Fix with a mechanical
> status boundary: TODO.md = open work (injected), DONE.md = completed archive
> (enforced + routed, NOT injected). Whole-file `cat` makes a second file a cleaner
> boundary than an in-file marker. Config: fold DONE.md under the existing `state`
> toggle (TODO + DONE + .local/todo are one unit — no new key). ADR 0009.

- [x] **E1 `DONE.md` scaffold + route-only spoke** (3f44dcb). New `templates/DONE.md` (archive
  header: "completed work; NOT injected; move `[x]` items here from TODO.md"). Added to
  `standard.Plan` under the `State` toggle. Added a hub route under State in
  `standard.Routes` — "Completed work / history → `DONE.md`". `hookPathsFor` never
  cats DONE.md — enforced + routed, never injected. TDD.
- [x] **E2 TODO.md = open-work-only + the move rule** (9174690). Updated `templates/TODO.md`
  header: open work only; "when a task is done, move its line to DONE.md — this file
  is injected every session, keep it small." Mirrored one state-hygiene convention
  line in `templates/docs/agents.md`. Reworded the State route.
- [x] **E3 `check` archive nudge (the enforcement half)** (f9a0113; threshold >5 `[x]`, soft
  note / exit 0). Heuristic in `gyroscope check`: flags when TODO.md carries more than
  5 completed `[x]` items → "archive done items to DONE.md". Soft note (prints, exits
  0) not drift — housekeeping shouldn't break CI. Threshold + severity documented in
  code + commit. TDD both branches.
- [x] **E4 SKILL.md reconciliation + adoption** (e526858). Extended the "reconcile, don't
  clobber" flow: on adoption, move existing `[x]` items to DONE.md and consolidate any
  stray/non-root TODO (e.g. `docs/TODO.md`) into root TODO.md — never two todo files.
- [x] **E5 ADR + docs + dogfood** (48a63b6). ADR 0009 (TODO/DONE split: mechanical status
  boundary beats the failed lean/detailed convention; DONE.md enforced + routed but
  not injected). Updated CONTEXT.md's Spoke + SessionStart-hook terms to include DONE.md.
  Dogfooded: split gyroscope's own TODO.md — moved every completed `[x]` block to DONE.md,
  left open items (E-series, Later) in TODO.md; refreshed the hub's managed region so it
  routes DONE.md; `./bin/gyroscope check .` stays conformant.

## Known non-issues (kept for context; not action items)
- [ ] (note, not fixed) `go build ./cmd/gyroscope` reports `dev (commit none, built unknown)` — no ldflags; `make build` is the versioned path. Expected, documented as the "quick local binary."

## Persona registration — mirror docs/agents → .claude/agents (2026-07-09)
> Problem (finding #2): `docs/agents/` personas are authored in valid Claude
> subagent format but live where Claude Code does NOT scan for subagents
> (`.claude/agents/`), so they never register as dispatchable agent types — the
> model reaches for registered ones (caveman `cavecrew`) instead. Fix (option C):
> keep `docs/agents/` canonical + hub-routed, and MIRROR each persona into
> `.claude/agents/` so Claude registers it. Claude-specific, enforcement-adapter-
> shaped, sits beside the SessionStart hook. Scope shift to note: the binary now
> COPIES persona files (registration) — it still never AUTHORS persona content
> (that stays the skill's job). Gate: `personas == on` AND `enforce.claude` on.
> Decisions: copy (not symlink — robust across clones/OS, drift caught by check);
> exclude `README.md` + any file without `name:` frontmatter; gyroscope owns the
> persona-named mirror files (overwrite on drift). Needs ADR 0010.

- [x] **F1 Persona mirror writer** (827e17d). New `internal/persona` package: when gated,
  copies each `docs/agents/*.md` that is a valid persona (has `name:` frontmatter) into
  `.claude/agents/<name>.md`, verbatim (dest filename from the frontmatter `name:`).
  Excludes `README.md` and frontmatter-less files. gyroscope owns the mirror files:
  atomic temp+rename overwrite on drift (the one deliberate non-WriteGuarded write).
  Missing/persona-less `docs/agents/` is a clean no-op. TDD: mirrors valid personas,
  excludes README + no-frontmatter files, dest bytes-equal source.
- [x] **F2 init wires the mirror** (879d1ab). `init --apply` registers personas by mirroring
  each into `.claude/agents/` when gated (personas on AND `enforce.claude`); dry-run lists
  the `.claude/agents/<name>.md` files it would write. Binary stays non-interactive; the
  mirror is a byte copy, not a render. Gate via `personaMirrorGated` (shared with check).
- [x] **F3 check verifies registration** (d6b11f9). When gated, for each canonical persona
  require `.claude/agents/<name>.md` present and byte-equal to its `docs/agents/` source
  (missing/differing = nonconformance); `check --fix` re-mirrors to converge. Extended the
  existing `PersonaOn` block in `check.go`. TDD: conformant when mirrored, drift when
  missing/edited, `--fix` re-mirrors.
- [x] **F4 Docs + ADR + dogfood** (1bf486c). ADR 0010 (persona registration mirror;
  the docs-vs-registered asymmetry vs cavecrew; copy-not-symlink; binary now mirrors persona
  bytes but still doesn't author content). Updated CONTEXT.md (refined "The binary" nuance +
  new "Persona mirror" term), the personas-spoke README template, and SKILL.md (personas now
  register into `.claude/agents/` via `init --apply`). Dogfooded on a FRESH notwhoop clone:
  after `agents set on` + `init --apply` (no --force), `.claude/agents/` holds all 7 personas
  byte-equal to their `docs/agents/` source, README not mirrored; gyroscope self-check stays
  conformant (no personas → no mirror required).

## Later — deferred (explicitly out of MVP)
- [ ] **PI persona sub-agents (extension-built) — parity for the `.claude/agents/` mirror.** PI has NO native sub-agent registry (by design — its README: "No sub-agents… build your own with extensions"), so there is nothing to mirror `docs/agents/` into the way F1–F4 do for Claude. Sub-agents are *buildable* via PI's extension API, though. To give PI parity, extend the gyroscope PI extension (`.pi/extensions/gyroscope-context.ts`) to register a sub-agent tool that, per `docs/agents/` persona, spawns `pi -p <prompt> --tools <persona tools>` with the persona as the system prompt. This is BUILDING runtime behavior (a harness feature), not a byte-copy mirror — larger scope than the Claude mirror, opt-in, gated on `enforce.pi` + `personas == on`. Keep `docs/agents/` canonical; the extension reads it live. See ADR 0010 (Claude mirror) for the docs-vs-registered rationale this would extend to PI.
- [ ] **Zed enforcement adapter** — active `enforce.Adapter` for Zed (the passive `.rules` doc-target pointer already exists; this is the force-inject side, parallel to Claude's SessionStart hook / PI's `session_start` extension). Investigate Zed's injection mechanism (does its agent support a session-start hook / rule that force-reads the hub, or is `.rules` native-read only?). If native-read only, there may be nothing to enforce — document that outcome. Wire behind the `enforce` config section (`zed`, opt-in) if a mechanism exists.
- [ ] **Cursor enforcement adapter** — active `enforce.Adapter` for Cursor (passive `.cursorrules` pointer already exists). Investigate Cursor's mechanism: `.cursor/rules/*.mdc` with `alwaysApply: true` (always-injected project rules) and/or Cursor hooks. An always-applied rule that force-reads the hub would be the enforcement analog. Wire behind the `enforce` config section (`cursor`, opt-in).
- [ ] **Native "terse" spoke (option c) — build if caveman-by-recommendation proves too weak.** Using caveman for now (see DONE.md, Post-MVP). The skill recommendation is discovery-only + setup-only; a gyroscope-owned terse spoke `cat` by the SessionStart hook would be *always-on*, dependency-free, and use the same mechanism caveman does (hook-injected rules) — no Node, no separate install, no skill nesting. Ship a static ruleset (drop filler/hedging/pleasantries; keep code, commands, and error strings byte-exact; **never** terse-ify security warnings or irreversible-action confirmations). Tradeoff vs caveman: no tuned levels / statusline / measurement. Wire as a `terse` `SpokeSet` toggle + `templates/docs/terse.md` + a `hookPathsFor` entry — mirror the `state` spoke. Context: caveman shrinks **output** tokens (~65%), leaves **input** untouched, so it never reduced the hook cost — it's the output-side complement to gyroscope's input-side docs.

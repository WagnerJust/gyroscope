# DONE

> Completed-work archive for this repo. Move a task's line here from `TODO.md`
> once it's finished. **This file is NOT injected** by the SessionStart hook — it
> is history, kept out of the per-session context so `TODO.md` stays a small
> "what's open" view. Tracked in git; append-only in spirit (edit for accuracy,
> don't prune the trail).

**Legend:** `[x]` done

## Cross-harness artifact sync — commands + MCP (2026-08-24)
> Design: `.local/design/2026-08-24-cross-harness-artifact-sync-design.md`. Solves the "dev uses
> neither gyroscope nor Claude Code" cases by extending the per-harness re-encode pattern to two new
> author-once artifact types. Scope: Claude/OpenCode/PI (enforce.* IS the harness list). **All phases
> done + LIVE-VALIDATED** against real opencode 1.18.15 + pi 0.84.3 (2026-08-24): OpenCode shows the
> generated persona as `reviewer (subagent)` with the tool-map `permission:` block, connects the
> merged MCP server (`opencode mcp list` → ✓ connected), and reads the injected `instructions`; PI
> registers the persona tool (`--tools reviewer` runs), expands the `/deploy` prompt-template, and —
> Phase 6 — the generated MCP-client extension loads, connects, registers every tool, and a tool call
> round-trips (MCP `inputSchema` → PI typebox `parameters` with no shim). `gyroscope check` stays
> conformant even after opencode rewrote its own `opencode.json` (validating the merge-preserve design).
- [x] **Phase 1 — cross-harness slash commands** (`38ad110`; ADR 0019). `internal/command` re-encodes
  `docs/commands/<name>.md` into `.claude/commands/`, `.opencode/commands/`, `.pi/prompts/` (body
  portable; frontmatter reduced per harness; Claude-only keys dropped). Presence-driven per enabled
  harness, independent of personas. `planMirrors` now unifies persona + command generated files → one
  path for init/check/sync.
- [x] **Phase 2 — OpenCode state-spoke injection fix** (`904fa47`). New `enforce.OpenCode` adapter
  merges the non-hub committed spokes into `opencode.json` `instructions` (the hook analog OpenCode
  lacked). Rides the adapter framework. `.local/*` excluded (documented). No ADR (extends ADR 0006).
- [x] **Phase 3 — cross-harness MCP** (`de4831e`; ADR 0020). Neutral `gyroscope.json mcp` block
  (`${VAR}` refs only) → `internal/mcp` transforms → Claude `.mcp.json` (dedicated, gyroscope-owned,
  rides `planMirrors`) + OpenCode `opencode.json mcp` (shared, merged by name). File-ownership splits
  by the tool's layout; shared-config no-prune trade documented.
- [x] **Phase 4 — sync bakes commands + dedicated MCP** (`0705045`). Wired free via `planMirrors`;
  e2e test proves a command file + `.mcp.json` reconstruct byte-for-byte on a fresh clone, `${VAR}`
  intact (no secret leak).
- [x] **Phase 5 — skill + docs**. SKILL.md documents cross-harness commands (+ adopt-existing), the
  MCP block (+ `${VAR}` safety + per-harness targets), and the OpenCode instructions injection.
  Decided AGAINST a hub route/scaffold for `docs/commands/` (generated-artifact feature, skill-discovered
  — keeps the standard lean).
- [x] **Phase 6 — PI MCP** (ADR 0021). PI has no native MCP, so gyroscope generates `.pi/mcp.json`
  PLUS its OWN minimal MCP-client extension `.pi/extensions/gyroscope-mcp/` (index.ts + package.json)
  using only the OFFICIAL `@modelcontextprotocol/sdk` (pinned) — NOT a community wrapper (Justin's
  call; consistent with dependency-light + the gyroscope-personas.ts precedent). All 3 files are
  dedicated owned mirrors → written/verified/baked-into-sync. Teammate runs `npm install` once. Stdio
  only for now. **Prototyped + validated live BEFORE encoding into Go** (the research's schema-bridging
  risk turned out to need no shim); re-validated from the real binary output. Low-regret: the config
  is the standard `mcpServers` shape a community extension also reads, so swappable.
- [x] **PI MCP remote transport** (finishing what was cut). The generated extension now connects
  **remote** streamable-HTTP servers (`url` + optional `headers`) as well as stdio (`command`) —
  validated live against the official "everything" server over http (`mcp_everything_echo` round-trips).
  Gotcha found + documented in ADR 0021: verify with `--tools <name>` (deterministic), NOT "list your
  tools" — the model formats that answer unreliably, which sent an earlier debug pass chasing a
  non-bug. `StreamableHTTPClientTransport` is statically imported (PI's loader compiles it fine).
- [x] **Relabel** the mirror-write message `mirrored persona X` → `generated X` (it now covers
  commands + MCP configs + extensions, not just personas).

## Now — lock the design
- [x] Agree thesis: opinionated, self-enforcing evolution of buckle's hub-and-spoke
- [x] Choose form: new standalone Go binary, `gyroscope`
- [x] Create repo home `~/Side/gyroscope` + `git init`
- [x] Draft design doc → `.local/design/2026-07-04-gyroscope-design.md`
- [x] Resolve open design questions (config format → planning detail)
  - [x] Hook mechanics: SessionStart + full hub & key spoke (token cost → caveman later)
  - [x] Harness-adapter shape: two registries (doc targets vs enforcement adapters)
  - [x] Interview home: skill drives grilling, binary writes (buckle pattern)
  - [x] Customization: repo config file (gyroscope.toml-style), idempotent
  - [x] Config format: JSON (`gyroscope.json`), stdlib-only
  - [x] Skill invocation: user-invoked (`disable-model-invocation`)
  - [x] Standard adds `docs/agents/`; `CONTEXT.md` seeded from interview
  - [x] Interview: the question set — 5 questions live in `skill/SKILL.md`
  - [x] The standard: concrete default spokes + templates → `templates/`
  - [x] Customization: DROP spokes via `gyroscope.json` ✓; ADD custom spokes via `custom` array — commit `60a6d86` (binary writes stub + injects route into the hub at a `<!-- gyroscope:custom-routes -->` marker)
  - [x] Distribution: `go install` + goreleaser (T9)
  - [x] Tool targets — MVP ships Claude + Gemini pointers; more deferred (see Later)
- [x] Draft the standard's templates + interview script → `.local/drafts/`
- [x] Reconcile drafted standard with plumbline ACMM rubric (draft-now-reconcile-later) — commits `ec0330f`+`92a0858`. Added the three L2 process artifacts (`.github/pull_request_template.md`, `.gitmessage`, `CONTRIBUTING.md`) as toggleable, default-on standard outputs; CONTRIBUTING is hub-routed and defers conventions to `docs/agents.md` (no drift), the other two are enforcement-genre (Git/GitHub apply them, no route). Verified: gyroscope now scores plumbline **L2 (Instructed) = 1.0** (was L1/0.25); all five L2 signals Found/NA. L3–L5 left to plumbline by design.
- [x] Implementation plan written → `.local/plans/2026-07-04-gyroscope-mvp-plan.md`
- [x] Final design sign-off → begin build (built, reviewed, dogfooded)

## Next — MVP build (after sign-off)
- [x] T1 Scaffold: `go.mod`, `embed.go`, templates, skill — commit `5801269`, build+vet green
- [x] T6 `gyroscope init` (non-interactive): answers (flags/spec) → write standard → install hook — `cmd/gyroscope`, commit `7b7cfd7`, combined review ✓ (+ exit-code fix, refuse-overwrite test, e2e verified)
- [x] T2 Config: `internal/config` — `gyroscope.json` toggles, default all-on — commit `151f8d1`, reviewed ✓
  - [x] follow-up (from review): check `WriteFile` err in test; malformed-JSON error-path test — commit `920d512`
- [x] T3 Standard writer: `internal/standard` — `Plan`+`Write` all spokes, `O_EXCL` guard, `.local/` gitignore safety (incl. `local.md` spoke) — commits `105b227`+`d497aaa`, spec✓ quality✓
- [x] T4 Doc-target registry + pointer-file writers (buckle-style) — `internal/target`, commit `7541c36`, combined review ✓
- [x] T5 Enforcement adapter interface + Claude adapter (writes `.claude/settings.json`) — `internal/enforce`, commit `1a89b49`, spec✓ quality✓ (+ hardening: atomic write, fail-loud on malformed hooks, no-HTML-escape output)
- [x] `gyroscope` skill: conversational interview → invokes the binary — ships as embedded `skill/SKILL.md` (T1) + `install-skill` command (T8)
- [x] T7 `gyroscope version` — `cmd/gyroscope/version.go`, commit `b4c9710`, reviewed ✓ (ldflags thread through, NoArgs→exit 2)
- [x] T8 Embed + install the gyroscope skill (`//go:embed`, per-tool targets) — `cmd/gyroscope/installskill.go`, commit `7df467e`, reviewed ✓ (fixed plan test's stray `skills/` path segment; installs to `<base>/gyroscope/SKILL.md`)
- [x] T9 Release: goreleaser + `go install` path — Makefile, `.goreleaser.yaml`, CI, README — commit `1d2932e`; `.goreleaser.yaml` later validated (`goreleaser check` + snapshot build ✓)
- [x] Tests — per-task TDD across all packages (config, standard, target, enforce, cmd); full suite green under `-race`

## Post-build review follow-ups (from final holistic review — all non-blocking; MVP shippable)
> None block shipping. Top of list = highest value. #1/#2/#4 involve a design call (yours).
- [x] (4f9cd19 → **superseded** 495d642) **#1 Dangling hub routes when a spoke is disabled** — first shipped as an honest hedge ("spokes are optional"). Now properly fixed: the binary renders the hub's routes from config, so disabled spokes are simply absent — no dead links, hedge removed. See ADR 0003.
- [x] (4f9cd19) **#2 `GEMINI` target registered but never written; `target.All()` is dead API** — `init` hard-codes `target.ByID("claude")` and never loops `All()`. Either loop `All()` to write every pointer (+ add `GEMINI.md` to the hub's pointer list) or drop the gemini entry + `All()` for MVP. **Design call: does gyroscope write GEMINI.md?**
- [x] (eb36719) **#4 Partial write leaves a hub with no enforcement; no clean recovery** — `init` returns on the first clobber before writing the pointer/hook; a repo with a pre-existing `AGENTS.md` gets nothing, and `--force` then clobbers curated content. Fix: pre-flight all destinations, fail before the first write → all-or-nothing `init`.
- [x] (4a931ce) #3 Duplicated clobber-guard write logic — `standard.Write` and `target.WritePointer` reimplement the same `MkdirAll`→`O_EXCL`/`O_TRUNC`→refuse-overwrite. Extract a shared `writeGuarded(dest, content, force)`.
- [x] (b40ed48) #6 Empty `gyroscope.json` → opaque `unexpected end of JSON input`; wrap the unmarshal error with the filename (`config.go:42`).
- [x] (df52737) #9 README omits `gyroscope.json` spoke toggles and `--force`; document them.
- [x] #7 `install-skill`: path-convention drift fixed — now takes a positional `[skills-dir]` matching `init` (commit `e199f0d`). Overwrite-on-apply kept by design (updating the managed skill should overwrite).
- [x] (b40ed48) #5 `exitInternal` (exit 4) is now unreachable (all errors route to exit 2 via `errCannotRun`) — route genuine internal failures to it or drop it.
- [x] (4f9cd19) #8 Pointer line says "routing **table**" but the hub uses a `## Routes` bulleted list — reword one.
- [x] (df52737) #10 goreleaser custom `ldflags` drop the default `-s -w` → release binaries aren't stripped (larger, not wrong).
- [x] (CI) gofmt gate added in CI (df52737). `.goreleaser.yaml` machine-validated: `goreleaser check` passes + `goreleaser build --snapshot` succeeds with correct ldflags. `dist/` gitignored.

## Dogfood findings (gyroscope run on itself — adoption commit 7a3c577)
- [x] (937a284) **SKILL never filled placeholders after `init`** — the headline promise (binary=structure, skill=content) was unimplemented. Added "Fill the scaffolds" + "Verify none remain" steps to `skill/SKILL.md`.
- [x] (937a284) **`{{...}}` marker collision** — reserved `{{...}}` for fill-once scaffolds only; ADR `TEMPLATE.md` now uses `<...>` per-use form fields; `embed_test.go` guards the invariant.
- [x] (ad59dca) Dry-run plan hides side effects — `init` appends `.local/` to `.gitignore`, and the hook `cat`s a gitignored personal file every session; neither is surfaced in the dry-run plan. Dry-run now prints both (the `.gitignore` line gated on the local spoke).
- [x] (7d2a9d7) Config-aware enforcement (hook) — `SessionStartCommand` is now a builder; `init` cats only the enabled spokes (`AGENTS.md` + `docs/agents.md` if agents-on + `.local/local.md` if local-on). The static hub still lists all routes, covered by the blanket "spokes are optional" hedge (review #1); per-route pruning would need the skill (binary doesn't template) — left as the hub's honest-hedge approach.
- [x] (752cf5e) `version` double-prints when untagged (`<sha> (commit <sha>, …)`) — `versionString` now collapses to `<sha> (built …)` when version == commit; used by both `version` and `--version`.

## Post-MVP — standard growth (Issue-driven, from "issues w/ agentic coding")
> Addressing the problem doc: agents don't follow docs (#1), setup/resumption is heavy (#2), output is wordy (#3).
- [x] L2 process artifacts — `.github/pull_request_template.md`, `.gitmessage`, `CONTRIBUTING.md` (commits `ec0330f`+`92a0858`); satisfies Issue #1 ("program docs into the tools") + plumbline L2.
- [x] **State file mandate (Issue #2 — resuming a new chat on existing work)** — the standard now writes a tracked, repo-wide `TODO.md` and a gitignored, personal `.local/todo.md`, both injected by the SessionStart hook so a fresh session resumes from current progress instead of re-deriving it. New `state` spoke (default on) in `internal/config`; `hookPathsFor` cats both; hub routes to `TODO.md`.
- [x] Issue #3 (wordy output → rubber-stamping) — **decided: rely on caveman for now.** gyroscope recommends it from the skill (`d89dd88`, opt-in step 6); caveman itself runs always-on via its own SessionStart hook on Claude Code (verified from `~/Src/caveman`) — no skill nesting, coexists with gyroscope's hook. A gyroscope-native "terse" spoke is deferred → see TODO.md Later. Do NOT bundle caveman's Node installer (breaks `go install`-clean).
- [x] Config-aware hub — the binary prunes routes for disabled spokes so the hub never dead-links (`495d642`, ADR 0003), superseding review #1's hedge.
- [x] Resume the ADR habit — ADR 0003 (config-aware hub) + ADR 0004 (standard scope: encoded judgement beyond docs) written after a gap since 0002. Keep writing them per the TEMPLATE bar.
- [x] Conformance / `check` command — `gyroscope check [repo]`: read-only inverse of init; verifies hub / routes==enabled-spokes / planned files present / pointer routing line / SessionStart hook / no unfilled `{{...}}`; exit 0 conformant / 1 drift / 2 can't-run. Reuses `Plan`/`Routes`/`hookPathsFor`/`target`/`enforce` (extracted `standard.Routes` + `enforce.HasSessionStart` to share, not duplicate). Scoped to conformance, NOT maturity (plumbline owns scoring). Automates dogfood pass 2; gyroscope checks conformant on itself.
- [x] plumbline audit-fit (coordinate with the bridge dev; plumbline owns auditing) — concrete work done: gyroscope scores **L2 = 1.0**; plumbline `main` recognizes the hub-and-spoke + pointer layout (the bridge dev's retrofit); the two recognition quirks are fixed (`1fdb070`, `488b47b`). Remainder is **standing coordination** — keep the two in sync as each evolves (new gyroscope spoke → plumbline still scores it; new plumbline L2 signal → gyroscope satisfies it) — not a discrete task.

## Dogfood pass 2 (2026-07-06 — after the standard grew process/state artifacts)
> Verdict: fresh `init` produces a coherent standard; repo is structurally faithful. All drift was in prose *describing* the standard — the exact thing gyroscope prevents — now fixed.
- [x] README under-described the standard — listed 5 spoke toggles, real `SpokeSet` has 9; intro omitted CONTRIBUTING / state / process artifacts. Fixed (docs commit).
- [x] `CONTEXT.md` stale — Spoke list, SessionStart-hook cat-set, and Scaffold definition predated the state/process artifacts (ADR 0004 had flagged this). Fixed + added a "Process artifact" term.
- [x] `skill/SKILL.md` omitted the state files. Fixed.
- [x] `docs/adr/TEMPLATE.md` drifted from the shipped template. Realigned byte-for-byte.

## More doc-target tools (2026-07-08)
- [x] **Cursor / Windsurf / Cline / Copilot / Zed pointers** — registered in `internal/target` (`.cursorrules`, `.windsurfrules`, `.clinerules`, `.github/copilot-instructions.md`, `.rules`), each writing the canonical routing line. One registry line each (ADR 0002 passive-pointer side); `All()`/`WritePointer`/`check` already loop it. Hub pointer list updated; gyroscope adopted the pointers on itself (check conformant). Skipped Aider (not auto-read without config) and Continue.

## PI enforcement adapter (2026-07-07)
- [x] **PI enforcement adapter** — `enforce.Adapter` interface extracted (`ID`/`PlanLine`/`Apply`/`Verify`); `PI{}` writes `.pi/extensions/gyroscope-context.ts` injecting the non-hub spokes on `session_start` via `pi.sendMessage` (hub read natively by PI, so excluded). New `enforce` config section (claude on, pi opt-in); init/check loop the enabled adapters; Claude wrappers keep its tested hook logic. Trust left to the user (`/trust`), not automated. ADR 0006. Branch `feat/pi-enforcement-adapter`.

## Persona wiring (config-driven — 2026-07-07)
- [x] **Config-driven persona wiring (`docs/agents/`)** — `spokes.personas` is now a four-state enum (`unknown|on|skipped|off`, back-compat with the bool); the hub carries a standing "ask when unknown" directive and the SessionStart hook cats `gyroscope.json` so the state is in context (hook stays pure `cat`). `gyroscope agents set <state>` records decisions; the skill reads a user-chosen template dir, customizes personas to this repo, and writes `docs/agents/*.md`. ADR 0005. Branch `feat/persona-wiring`.

## DX convergence — merge-safe init + fix loop (2026-07-08)
> Goal: kill the "refuse all-or-nothing vs `--force` clobber" choice and the
> "skill or binary?" confusion. Make `init` idempotent + merge-safe, expose a fix
> loop, unify the two tools. Design rationale: conversation 2026-07-08. Dissolves
> the buckle-migration pain, the custom-route-trips-`check` edge, and the collision
> refuse in one design move. Write an ADR for the managed-block standard change.

- [x] **D1 Per-file convergence classifier.** `init` dry-run classifies each
  destination: NEW / OK (present & conformant) / MERGE (present, missing managed
  content) / CONFLICT (user content differs). Print per-file status instead of the
  all-or-nothing collision refuse. Evolve `existingCollisions` (`init.go:154`) into
  a classifier. TDD. — `cmd/gyroscope/converge.go` classifier +
  `internal/standard/managed.go` region primitives (`MergeManaged`); dry-run now
  prints per-file state; hub template wrapped in `<!-- gyroscope:managed -->`.
  `existingCollisions` retired for `preexisting`/`conflicts`. Commit `99a53c3`.
  Follow-up (migration case): a hand-written hub with NO managed markers now
  classifies MERGE (D1's "present, missing managed content"), not CONFLICT —
  `MergeManaged` appends the wrapped region at EOF, preserving all user prose above
  it. Verified against a real 58-line hand-written hub (notwhoop).
- [x] **D2 Managed-block boundary for the hub.** Generalize the existing
  `<!-- gyroscope:custom-routes -->` marker to a full managed region in `AGENTS.md`:
  gyroscope owns only content between `<!-- gyroscope:managed -->` /
  `<!-- /gyroscope -->`; everything outside is the user's — untouched, invisible to
  `check`. Fixes buckle merge + custom-routes edge + idempotent re-init. Update
  `check`'s Routes comparison to read only the managed region. **Needs an ADR** —
  breaking change to the standard's hub format. — `check` now extracts the managed
  region first (route/personas checks scoped to it; user content outside is
  invisible; a hub with no region is drift). Repo's own `AGENTS.md` migrated to the
  managed-block form. ADR 0007. New "Managed region" term in `CONTEXT.md`. Commit
  `2d9b21c`.
- [x] **D3 `init --apply` merge-safe.** Apply the NEW + MERGE subset automatically
  (create missing files; inject missing managed content into an existing hub). Only
  a true CONFLICT needs `--force`. Whole-file writes keep `fsutil.WriteGuarded`;
  in-place managed-block injection is a new merge path (still atomic temp+rename).
  — `applyConverge` drives writes per classified item (shared by init and D4's
  fix); `standard.InjectManaged` is the atomic temp+rename managed merge (the one
  deliberate WriteGuarded exception, justified in a comment). init is now
  idempotent; re-apply on a current repo writes nothing. Commit `35cfe88`.
  Follow-up (mixed repo): `--apply` (no `--force`) is no longer all-or-nothing — it
  now writes every NEW + MERGE item and SKIPS only the CONFLICTs, printing
  `N conflict(s) skipped (use --force): …` and exiting drift (1) so the remaining
  conflict is visible (0 when fully converged). `--force` still overwrites. Verified
  against notwhoop: the NEW spokes land even though CLAUDE.md conflicts.
- [x] **D4 `check --fix` (or `init --fix`).** Auto-apply the safe convergence so
  `check` (detect) and fix (converge) are symmetric. CI runs `check`; dev runs
  `--fix`. — `check --fix` runs the shared `applyConverge` in skip-conflicts mode
  (create NEW, merge the hub's managed region), then re-checks; conflicts are never
  clobbered and still surface as drift (exit 1). README documents both. Commit
  `d135ba1`.
- [x] **D5 Unify binary + skill.** `install-skill` guarantees the binary is
  resolvable (warn + install instructions when `gyroscope` isn't on PATH; skill
  step 2 shells to it). Removes the "skill installed but binary absent → step 2
  fails" trap. — `install-skill` now resolves `gyroscope` via `exec.LookPath`
  (injectable `lookBinary` for tests): confirms the path when present, warns with a
  `go install` instruction when absent (still installs the skill). SKILL.md step 2
  documents the dependency + the merge-safe apply (final commit on branch
  `feat/dx-convergence`).
- Out of scope (deferred): full 3-way content merge (managed blocks get ~90% for
  ~10% of the code — build only if blocks prove too weak). The binary stays
  non-interactive; interactivity lives in the skill.

## Skill: model-invocable + existing-repo reconciliation (2026-07-08)
> Reverses the deliberate "user-invoked" decision — see ADR 0008.
- [x] **Flip `/gyroscope` to model-invocable** — removed `disable-model-invocation`
  from `skill/SKILL.md`; broadened the description so NL ("get this repo up to date
  with gyroscope", adopt, migrate a buckle hub, or fresh setup) triggers it.
  Strengthened the HARD-GATE (now load-bearing: auto-invocation never bypasses
  show-plan-then-approve). Updated CONTEXT.md's "skill" term. ADR 0008.
- [x] **Existing-repo reconciliation flow** — added a "reconcile, don't clobber"
  section to `skill/SKILL.md`: `check` → dry-run classify → `init --apply` (no
  force) → hand-reconcile each CONFLICT (pointer content relocation, artifact
  keep-vs-adopt, hub route dedup) → fill from existing docs → `check --fix` to green.
  Closes the gap the notwhoop live test exposed (SKILL.md was fresh-scaffold-only).

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

## Archive nudge convergence — `check --fix` archives done items (2026-07-10)
- [x] **TODO→DONE move mechanized into `check --fix`.** The archive nudge (ADR 0009) fired at
  check time to whoever ran check — usually the human — never reaching the mid-session agent
  that finished a task, so the "move done items" rule needed manual reminding despite being
  stated 3×. Now `check --fix` moves completed *top-level* `[x]` items (with their indented
  sub-trees) out of the injected `TODO.md` into `DONE.md`'s `## Completed` section (newest on
  top), converging the only check finding `--fix` previously couldn't fix. New pure
  `internal/archive` package (`Plan`/`Merge`); new `fsutil.WriteAtomic` (temp+rename overwrite,
  the second deliberate non-`WriteGuarded` write after the persona mirror), DONE-written-first
  so a partial failure duplicates rather than loses work. Gated on `--fix` + the `state` spoke;
  no-op when nothing is done. ADR 0011. Rejected a 4th prose line, a Stop-hook nag, and an
  auto-mutating SessionStart hook (see ADR 0011 for why). Template `TODO.md` header updated to
  point at the mechanized path.

## Contributor-facing, tool-optional docs (2026-07-10)
- [x] **gyroscope-maintained repos are first-class for devs without gyroscope.** Generalized the
  managed-region mechanism from hub-only to any spoke: `standard.InjectManaged` takes a `dest`,
  converge's classifier keys the MERGE path on "want has a managed region" (not `dest ==
  "AGENTS.md"`), and `check` byte-verifies each non-hub managed spoke (hub stays semantically
  checked — its region is config-rendered). Added a managed **contributor block to
  `CONTRIBUTING.md`** ("you don't need gyroscope installed; what the markers/`gyroscope.json`/
  `.local/` mean; hand-edits reconcile on the next `check --fix`; optional zero-install `go run
  …@latest check .`") plus a one-line agent-facing pointer in the hub's managed region. Existing
  adopters converge via `MergeManaged`'s markerless-append path (user prose preserved); `check
  --fix` performs it. ADR 0012; CONTEXT.md "Managed region" term generalized. Dogfooded on this
  repo (hub line + block merged, re-check conformant).

## `.local/local.md` ready-to-use — no unfillable placeholder (2026-07-10)
- [x] **Fixed the `.local` placeholder papercut.** `check --fix` created `.local/local.md` from a
  `{{...}}`-marker scaffold that only the skill fills — but the skill never fills it (it is
  personal, machine-specific, dev-authored). So on a fresh clone `--fix` manufactured a placeholder
  the binary could not clear, regressing `check` from green (absence tolerated) to unfixable drift.
  Root cause was miscategorization: `.local/local.md` is a *ready-to-use* personal file like
  `.local/todo.md`, not a skill-fill scaffold. Rewrote `templates/local.md` as ready-to-use plain
  prose (no `{{...}}`), so `check` no longer polices it and `--fix` converges a fresh repo to
  conformant. Regression test `TestLocalScaffoldIsReadyToUse`. Verified on a fresh exodus worktree:
  `check --fix` → conformant. Found while converging the exodus branch with the post-release binary.

## Contributor block was truncated by self-referential markers (2026-07-10)
- [x] **Fixed the truncated CONTRIBUTING.md contributor block.** The block's prose explained the
  markers by quoting them literally (`<!-- gyroscope:managed -->` / `<!-- /gyroscope -->`).
  `ManagedRegion` locates a region by the *first* close marker after the open, so it stopped at the
  in-prose marker and shipped a block cut off mid-bullet — missing the `gyroscope.json`/`.local`
  notes, the "you never have to run gyroscope" paragraph, and the zero-install command. Invisible to
  `check` because `want` and `got` truncate identically (byte-equal → conformant), and the original
  tests only asserted early substrings. Fix: rewrite the prose to name the markers without the raw
  `<!-- ... -->` syntax. Guards: `TestManagedTemplatesHaveNoNestedMarkers` (no managed template's
  region may embed a marker literal) + strengthened the converge test to assert the block's tail
  (zero-install line + well-formed close) lands. Re-converged this repo; exodus + any adopter
  self-heal on the next `check --fix` with the fixed binary. Caught by Justin eyeballing the exodus
  CONTRIBUTING.md. (Deeper option not taken: line-anchor marker detection in `ManagedRegion`.)

## Skill adopt-reconcile step for pre-existing personas (2026-07-10)
- [x] **Persona flow now reconciles what's already on disk before wire/skip.** The flow was
  greenfield-only (ask → pick templates → write → mirror) with no branch for a repo that already
  has personas; `gyroscope check` is config-relative, so a `skipped`/`unknown`/`off` state never
  inspects `docs/agents/` and green said nothing about unregistered persona files (esp. in a
  `docs/agents/personas/` subdir the root-only mirror can't see). Added a **first** step to
  `skill/SKILL.md`'s persona flow (commit `beb7516`): scan `docs/agents/**` recursively for
  `name:`-bearing files regardless of config state (`grep -rl '^name:'`), classify root-canonical
  (ADR 0010) vs subdir (ask which — doc-routed-on-purpose vs mislaid, don't assume), reconcile
  state-vs-disk disagreement, then fall through to wire/skip. **Binary untouched** — it stays
  config-relative and deterministic; the judgment lives in the skill (the binary/skill split).
  Motivated by the exodus adopt session; see the doc-routed-persona pattern in
  `review-checkpoints.md` and ADR 0010. (TODO line and the SKILL.md step landed in the same commit;
  archived here once verified present in HEAD.)

## Adopter hardening — PR #5 (2026-07-10)
- [x] **Adopter-facing scaffold fixes.** `templates/docs/agents/README.md` cited `docs/adr/0010` (an
  internal gyroscope ADR no adopter scaffolds) — dead link removed. `templates/CONTRIBUTING.md`
  asserted "a maintainer or CI keeps the docs conformant" — gyroscope writes no CI — reworded to
  conditional. Commit `ad4390d`. Found in a cold-reader review of an adopted repo.
- [x] **`check` warns when `.claude/` is gitignored** (`d8cdbcf`). The enforcement hook
  (`.claude/settings.json`) and persona mirror (`.claude/agents/`) live under `.claude/`; a blanket
  gitignore keeps them out of version control, so teammates get the tracked docs but no
  auto-injection and no dispatchable subagents. Soft note (advisory, exits 0) advising a targeted
  `!.claude/settings.json` negation; heuristic `.gitignore` parse, no git dependency.
- [x] **`enforce.aiAttribution` toggle** (`c3e82cb` + `00ceb07`). New `gyroscope.json` key (default
  `true`). When `false`: Claude-native — `init`/`check --fix` write `includeCoAuthoredBy:false` into
  `.claude/settings.json` (drops the `Co-Authored-By` trailer + generated-by line); cross-harness — a
  managed hub directive (SessionStart-injected) tells PI/Cursor/etc. and reinforces Claude to add no
  attribution/sign-off. `check` verifies both. Scoped to AI credit, **not** the developer's DCO
  `Signed-off-by`. Works for non-gyroscope teammates via the committed settings + injected rule.
  Refactored settings.json read/atomic-write into shared `readSettings`/`writeSettings`.
- [x] **Cold-reader legibility (scoped slice)** (`afc6af5`). Added a "what these files are" map to the
  CONTRIBUTING managed block — pointer files (redirects, not config), `docs/agents.md` (file) vs
  `docs/agents/` (dir), and that `.gitmessage` needs `git config commit.template` to activate.
  Explain, don't restructure; structural changes (renames, per-file comments) left by design.

## Enforce ADR usage + upkeep (2026-07-23)
- [x] **ADR enforcement — directive (write them) + structural check (keep them healthy).** Mirrors
  the personas/attribution pattern, gated on the `adr` spoke. **Usage:** a standing managed-hub
  directive (`## Architecture decisions`, new `<!-- gyroscope:adr-directive -->` marker) instructs
  agents to record significant decisions as ADRs and echoes the TEMPLATE's own three-part guard so it
  encourages without inflating the count; `standard.ADRDirective` is the single source of truth,
  verified by `check`. **Upkeep:** new pure `internal/adr` package checks contiguous numbering from
  0001, a `**Status:**` + `**Date:**` on every record, and no reference to a nonexistent ADR number
  (the class of bug behind the shipped `docs/adr/0010` dead link); TEMPLATE.md excluded so its
  `ADR-XXXX` placeholders don't register. Resolved the open notes-vs-drift question: structural
  findings are **drift** (fail CI, like the dangling-route check), the directive stays soft. Rationale
  in ADR 0013. TDD throughout; dogfooded (`check .` green with the hub converged + ADR 0013 added).

## Harness coverage — AGENTS.md-native cohort (2026-07-23)
- [x] **Codex + OpenCode acknowledged as native `AGENTS.md` readers.** Verified against official
  docs (2026-07-23): OpenAI Codex CLI reads `AGENTS.md` automatically every session (global →
  root→cwd, closer-wins, 32 KiB cap) and OpenCode reads it by default (first-match-wins); **neither
  needs a pointer file, config entry, or session hook** — nothing to wire. Conclusion is
  documentation-only: a `target` registry entry would be *wrong* (it generates a redundant pointer).
  Added a one-line acknowledgment to the hub's "Pointer files" section (ships in the standard for
  every adopter) naming both as native consumers, plus a regression test. Lean into the
  strength: the hub already speaks the format these tools read. (Cursor/Zed active adapters remain
  separate, deferred.)
- [x] **Overclaim corrected** (`79ed6a7`). The above initially called Codex/OpenCode "first-class here
  with nothing extra to install" — false: reading the hub natively covers only *injection*, not the
  rest of harness parity (persona-subagent registration + attribution). Reworded to state only what is
  true (no pointer/hook needed to inject the hub). "Ready" = the full Claude triad, built below.
- [x] **OpenCode + Codex persona registration** (`f779da6`; ADR 0014). Extended the Claude persona
  mirror (ADR 0010) to both harnesses' native, project-committable subagent registries. Byte-copy no
  longer sufficed (formats differ), so the binary now RE-ENCODES gyroscope-owned mirror files:
  `persona.Target{dest,render}`; `PlanFor` parses once and emits Claude `.claude/agents/*.md`
  (byte-copy), OpenCode `.opencode/agents/*.md` (frontmatter → `description` + `mode: subagent`, body
  verbatim), Codex `.codex/agents/*.toml` (body → `developer_instructions`, escaped). New opt-in
  `enforce.opencode`/`enforce.codex` toggles, gated on `personas==on`. check verifies each mirror
  byte-for-byte; `check --fix` re-encodes. Ruled the binary-re-encode-vs-skill fork toward the binary
  (mechanical re-encode of owned files ≠ ADR-0003 content-templating).
- [x] **Codex attribution** (`bef4cc2`). `enforce.codex` + `aiAttribution:false` writes top-level
  `commit_attribution=""` into `.codex/config.toml` (Codex's co-author kill-switch); check flags a
  missing/non-empty value. Line-edit (no TOML dependency): replace/prepend/create, preserving other
  content, atomic. OpenCode has no attribution setting → cross-harness hub directive only.
- [x] **PI persona parity** (`622fbcb`; ADR 0015). PI has no native registry, so parity is a generated
  runtime extension `.pi/extensions/gyroscope-personas.ts` registering one subagent tool per persona
  (`pi.registerTool`), each spawning headless `pi -p --append-system-prompt <body>`. gyroscope bakes
  the persona metadata (single parser, no TS-side YAML), rides the mirror pipeline (one owned file,
  verified byte-for-byte). Tool-restriction mapping (`--tools`) is a documented gap (Claude names
  don't map to PI's). Confirmed against PI's extension API + headless flags before building. Full
  harness matrix (Claude/PI/OpenCode/Codex) now reaches persona-dispatch parity when its adapter is on.
- [x] **Persona tool-allowlist carries through (tool-fidelity follow-up; ADR 0015 update 2026-07-24).**
  Closed the gap the parity build flagged: the persona `tools:` allowlist is now mapped onto each
  harness's native restriction model, verified against each tool's docs/source first. OpenCode — a
  per-tool `permission:` block (allow the persona's tools, `deny` the rest of the mappable set; `deny`
  is a hard block). PI — `--tools` with the mapped built-in names (Read→read, Glob→find, …); only
  verified names are baked, so an unknown name can never fail the run; WebFetch/WebSearch dropped (no
  PI built-in). Codex — `sandbox_mode = "read-only"` when a persona touches no mutating/network tool,
  else left at default (only tightens, never loosens; Codex has no per-tool allowlist). parse() now
  captures the allowlist; maps live in `internal/persona/toolmap.go`. Residual low-priority gaps
  (web tools, Codex coarseness, verify-on-target) tracked in TODO. TDD.

## Cursor + Zed enforcement (2026-07-24; ADR 0016)
- [x] **Cursor enforcement adapter** (`enforce.cursor`, opt-in). Verified Cursor has a real
  force-inject primitive: a `.cursor/rules/*.mdc` project rule with `alwaysApply: true` is injected
  into every request. New `enforce.Cursor` adapter writes `.cursor/rules/gyroscope-hub.mdc`
  (alwaysApply + "read AGENTS.md and follow its routes"), rides the standard adapter interface (init
  applies, check verifies, --fix restores). Static rule (a Cursor rule can't cat live files, so it
  force-injects the *directive*, not hub contents; the hub routes onward). Supersedes the passive,
  deprecated `.cursorrules` pointer for agent-mode Cursor while keeping it for back-compat. TDD.
- [x] **Zed enforcement adapter — resolved as a no-op (documented).** Verified Zed has no
  hook/always-apply flag; its Agent auto-reads the first-match instruction file (`.rules` wins over
  `AGENTS.md`) on every session. So gyroscope's existing passive `.rules` pointer IS the mechanism —
  nothing to build. Documented in ADR 0016 (with the first-match caveat: keep `.rules` pointing at the
  hub). Enforcement matrix now: active adapter for Claude/PI/Cursor; native auto-read for
  Zed/OpenCode/Codex; passive pointer for the rest.

## Sync script — reconstruct mirrors for a fresh clone (2026-07-24; ADR 0018)
- [x] **`sync` toggle — committed, idempotent reconstruction script.** The second answer to the
  ship-to-team gap (besides committing the generated dirs): opt-in `gyroscope.json` `sync` flag
  generates `gyroscope-sync.sh`, which rebuilds every enabled harness's persona-registration mirror on
  a fresh clone. **Bakes rendered bytes (base64), not a copy-list** — bash can't reproduce the
  OpenCode/Codex/PI re-encodes without a drifting second encoder, so it carries the encoders' *output*;
  regenerated from the mirror plan on init/check --fix, verified byte-for-byte by check. **Idempotent**
  (skip-if-exists → "sync if it hasn't been", never clobbers a dev's own agent). **Opt-in + dev-run,
  not hook-wired** — the toggle is the consent; the pure-`cat` SessionStart hook is left untouched
  (no read→write change, no auto-run trust surface); hook auto-invocation is a future add-on behind
  the same toggle. POSIX sh + base64 (Windows gap documented). Scope: the mirrors, not the Claude
  settings hook (JSON merge unsafe to bake — commit `settings.json` instead). Proven by an e2e test
  that runs the script and checks byte-for-byte reconstruction + idempotency. ADR 0018.

## Code-review persona preset (2026-07-24; ADR 0017)
- [x] **`codeReview` preset — opt-in curated review-suite scaffold.** Resolved the toggle-vs-preset
  question: a single `gyroscope.json` `codeReview` flag (default off) that REDIRECTS the persona flow
  — the binary only records the flag (one config field, no templates; binary/skill split holds), the
  `/gyroscope` skill scaffolds the bundle. SKILL.md now carries the canonical suite (distilled from
  `~/Work/lajolla`): layer-aligned `*-reviewer.md` specialists (repo-tuned — a Go CLI gets
  backend+security+cicd, drops web genres), `security-reviewer` always-on, a non-filtering
  `synthesizer`, and the load-bearing `intent-verifier` that grounds findings against the repo's
  decision record (ADRs/TODO/DONE/spokes), plus a `code-review` orchestrator runbook and a grounding
  ADR recorded in the adopting repo. Reviewer personas are ordinary personas → register into every
  enabled harness registry, policed by existing checks (no new enforcement surface). Complements a
  generic runner (Claude `/code-review`); this supplies the repo's own reviewers + intent-grounding.
- [x] **SKILL.md brought current with the harness-parity work.** The skill's persona-registration and
  enforcement docs only mentioned `.claude/agents/` + PI context injection — stale after the parity
  build. Updated: the register step now documents per-harness registries (Claude byte-copy;
  OpenCode/Codex re-encode; PI personas extension) gated on `enforce.*` and the `tools:` carry-through;
  the enforcement section now covers all toggles (claude/pi/opencode/codex/cursor/aiAttribution) with
  Zed's native-read no-op noted — so the shipped features are discoverable through the skill.

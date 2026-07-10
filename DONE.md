# DONE

> Completed-work archive for this repo. Move a task's line here from `TODO.md`
> once it's finished. **This file is NOT injected** by the SessionStart hook — it
> is history, kept out of the per-session context so `TODO.md` stays a small
> "what's open" view. Tracked in git; append-only in spirit (edit for accuracy,
> don't prune the trail).

**Legend:** `[x]` done

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

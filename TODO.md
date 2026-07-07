# gyroscope — build TODO

> Living build tracker, hand-maintained (not via the todo tool).
> Design docs & scratch live in `.local/` (gitignored). This file is tracked.

**Legend:** `[ ]` not started · `[~] `in progress · `[x]` done

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
- [x] Issue #3 (wordy output → rubber-stamping) — **decided: rely on caveman for now.** gyroscope recommends it from the skill (`d89dd88`, opt-in step 6); caveman itself runs always-on via its own SessionStart hook on Claude Code (verified from `~/Src/caveman`) — no skill nesting, coexists with gyroscope's hook. A gyroscope-native "terse" spoke is deferred → see Later. Do NOT bundle caveman's Node installer (breaks `go install`-clean).
- [x] Config-aware hub — the binary prunes routes for disabled spokes so the hub never dead-links (`495d642`, ADR 0003), superseding review #1's hedge.
- [x] Resume the ADR habit — ADR 0003 (config-aware hub) + ADR 0004 (standard scope: encoded judgement beyond docs) written after a gap since 0002. Keep writing them per the TEMPLATE bar.

## Dogfood pass 2 (2026-07-06 — after the standard grew process/state artifacts)
> Verdict: fresh `init` produces a coherent standard; repo is structurally faithful. All drift was in prose *describing* the standard — the exact thing gyroscope prevents — now fixed.
- [x] README under-described the standard — listed 5 spoke toggles, real `SpokeSet` has 9; intro omitted CONTRIBUTING / state / process artifacts. Fixed (docs commit).
- [x] `CONTEXT.md` stale — Spoke list, SessionStart-hook cat-set, and Scaffold definition predated the state/process artifacts (ADR 0004 had flagged this). Fixed + added a "Process artifact" term.
- [x] `skill/SKILL.md` omitted the state files. Fixed.
- [x] `docs/adr/TEMPLATE.md` drifted from the shipped template. Realigned byte-for-byte.
- [ ] (note, not fixed) `go build ./cmd/gyroscope` reports `dev (commit none, built unknown)` — no ldflags; `make build` is the versioned path. Expected, documented as the "quick local binary."

## Later — deferred (explicitly out of MVP)
- [ ] plumbline audit-fit (coordinate; another dev owns the bridge)
  - Two plumbline-side quirks surfaced during the ACMM reconciliation (fix in plumbline, not gyroscope): (1) `GEMINI.md` is not in plumbline's recognized agent-instruction paths, so a Gemini-only hub consumer is invisible to the L2 scan; (2) plumbline's `nextGap` filters on `score < Found` and ignores `NA`, so it can list our `l2.instructions-no-drift` (NA by design) as a "gap to fix."
- [ ] **Native "terse" spoke (option c) — build if caveman-by-recommendation proves too weak.** Using caveman for now (see Post-MVP). The skill recommendation is discovery-only + setup-only; a gyroscope-owned terse spoke `cat` by the SessionStart hook would be *always-on*, dependency-free, and use the same mechanism caveman does (hook-injected rules) — no Node, no separate install, no skill nesting. Ship a static ruleset (drop filler/hedging/pleasantries; keep code, commands, and error strings byte-exact; **never** terse-ify security warnings or irreversible-action confirmations). Tradeoff vs caveman: no tuned levels / statusline / measurement. Wire as a `terse` `SpokeSet` toggle + `templates/docs/terse.md` + a `hookPathsFor` entry — mirror the `state` spoke. Context: caveman shrinks **output** tokens (~65%), leaves **input** untouched, so it never reduced the hook cost — it's the output-side complement to gyroscope's input-side docs.
- [ ] Agency-persona wiring (`docs/agents/`)
- [ ] Conformance / `check` command
- [ ] PI coding tool enforcement adapter (next harness after Claude)
- [ ] More doc-target tools beyond Claude Code

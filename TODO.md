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
  - [ ] Interview: the question set `gyroscope init` asks
  - [ ] The standard: concrete default spokes + templates
  - [ ] Customization: how users add/drop spokes
  - [ ] Distribution: `go install` / goreleaser / skills.sh?
  - [ ] Tool targets beyond Claude Code
- [x] Draft the standard's templates + interview script → `.local/drafts/`
- [ ] Reconcile drafted standard with plumbline ACMM rubric (draft-now-reconcile-later)
- [x] Implementation plan written → `.local/plans/2026-07-04-gyroscope-mvp-plan.md`
- [ ] Final design sign-off → begin build

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
- [x] (4f9cd19) **#1 Dangling hub routes when a spoke is disabled** — `templates/AGENTS.md` is static and always lists all 5 routes, but disabling a spoke (e.g. `{"spokes":{"agents":false}}`) skips writing `docs/agents.md`, leaving a dead link. Options: conditionally template the routes, hedge every route with "may not exist" (only `.local/` does today), or have the skill prune the hub. **Design call.**
- [x] (4f9cd19) **#2 `GEMINI` target registered but never written; `target.All()` is dead API** — `init` hard-codes `target.ByID("claude")` and never loops `All()`. Either loop `All()` to write every pointer (+ add `GEMINI.md` to the hub's pointer list) or drop the gemini entry + `All()` for MVP. **Design call: does gyroscope write GEMINI.md?**
- [x] (eb36719) **#4 Partial write leaves a hub with no enforcement; no clean recovery** — `init` returns on the first clobber before writing the pointer/hook; a repo with a pre-existing `AGENTS.md` gets nothing, and `--force` then clobbers curated content. Fix: pre-flight all destinations, fail before the first write → all-or-nothing `init`.
- [x] (4a931ce) #3 Duplicated clobber-guard write logic — `standard.Write` and `target.WritePointer` reimplement the same `MkdirAll`→`O_EXCL`/`O_TRUNC`→refuse-overwrite. Extract a shared `writeGuarded(dest, content, force)`.
- [x] (b40ed48) #6 Empty `gyroscope.json` → opaque `unexpected end of JSON input`; wrap the unmarshal error with the filename (`config.go:42`).
- [x] (df52737) #9 README omits `gyroscope.json` spoke toggles and `--force`; document them.
- [~] #7 `install-skill` overwrites unconditionally — WON'T FIX (by design: `install-skill --apply` should update the gyroscope-managed skill to the latest; overwrite is correct there). `--dir` vs positional path drift left as-is.
- [x] (b40ed48) #5 `exitInternal` (exit 4) is now unreachable (all errors route to exit 2 via `errCannotRun`) — route genuine internal failures to it or drop it.
- [x] (4f9cd19) #8 Pointer line says "routing **table**" but the hub uses a `## Routes` bulleted list — reword one.
- [x] (df52737) #10 goreleaser custom `ldflags` drop the default `-s -w` → release binaries aren't stripped (larger, not wrong).
- [x] (CI) gofmt gate added in CI (df52737). `.goreleaser.yaml` machine-validated: `goreleaser check` passes + `goreleaser build --snapshot` succeeds with correct ldflags. `dist/` gitignored.

## Dogfood findings (gyroscope run on itself — adoption commit 7a3c577)
- [x] (937a284) **SKILL never filled placeholders after `init`** — the headline promise (binary=structure, skill=content) was unimplemented. Added "Fill the scaffolds" + "Verify none remain" steps to `skill/SKILL.md`.
- [x] (937a284) **`{{...}}` marker collision** — reserved `{{...}}` for fill-once scaffolds only; ADR `TEMPLATE.md` now uses `<...>` per-use form fields; `embed_test.go` guards the invariant.
- [ ] Dry-run plan hides side effects — `init` appends `.local/` to `.gitignore`, and the hook `cat`s a gitignored personal file every session; neither is surfaced in the dry-run plan.
- [ ] Config-aware enforcement — `SessionStartCommand` statically `cat`s all spoke paths even when a spoke is toggled off (degrades via `2>/dev/null`); static hub still lists disabled-spoke routes (overlaps review #1). Make the hub + hook config-aware.
- [ ] `version` double-prints when untagged (`<sha> (commit <sha>, …)`) — collapse to one value or label `dev-<sha>`.

## Later — deferred (explicitly out of MVP)
- [ ] plumbline audit-fit (coordinate; another dev owns the bridge)
- [ ] Token reduction (evaluate `caveman`)
- [ ] Agency-persona wiring (`docs/agents/`)
- [ ] Conformance / `check` command
- [ ] PI coding tool enforcement adapter (next harness after Claude)
- [ ] More doc-target tools beyond Claude Code

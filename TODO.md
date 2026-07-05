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
- [ ] Scaffold: `go.mod`, `cmd/gyroscope`, `Makefile`, CI
- [ ] `gyroscope init` (non-interactive): answers (flags/spec) → write standard → install hook
- [ ] Config model + reader (`gyroscope.toml`-style; spokes on/off + custom)
- [ ] `local.md` spoke: write `.local/local.md` + add `.local/` to target `.gitignore`
- [ ] Doc-target registry + pointer-file writers (buckle-style)
- [ ] Enforcement adapter interface + Claude adapter (writes `.claude/settings.json`)
- [ ] `gyroscope` skill: conversational interview → invokes the binary
- [ ] `gyroscope version`
- [ ] Embed + install the gyroscope skill (`//go:embed`, per-tool targets)
- [ ] Release: goreleaser + `go install` path
- [ ] Tests

## Later — deferred (explicitly out of MVP)
- [ ] plumbline audit-fit (coordinate; another dev owns the bridge)
- [ ] Token reduction (evaluate `caveman`)
- [ ] Agency-persona wiring (`docs/agents/`)
- [ ] Conformance / `check` command
- [ ] PI coding tool enforcement adapter (next harness after Claude)
- [ ] More doc-target tools beyond Claude Code

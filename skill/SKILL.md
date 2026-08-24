---
name: gyroscope
description: Set up OR bring a repo up to date with gyroscope's opinionated, self-enforcing agent-doc standard — the `AGENTS.md` hub, its spokes, the pointer files, and the enforcement hook. Use when the user wants to adopt gyroscope, "get this repo up to date with gyroscope", make `gyroscope check` pass, migrate an existing or buckle-style hub, or scaffold agent docs from scratch. Interviews only where the repo can't answer, and always gets approval before writing.
---

# gyroscope

Set up gyroscope's standard, or bring an existing repo up to date with it: an
opinionated hub-and-spoke doc set (`AGENTS.md` hub, `CONTEXT.md` glossary,
`docs/agents.md` instructions, `docs/adr/`, `docs/agents/` personas), the L2 process
artifacts (`CONTRIBUTING.md` routed from the hub, plus `.github/pull_request_template.md`
and `.gitmessage`, which Git/GitHub apply directly), the state files (`TODO.md` +
gitignored `.local/todo.md`), and a `SessionStart` hook that injects the hub +
instructions + current state every session, so a fresh chat resumes where the last
one stopped. The `gyroscope` binary does the writing — this skill gathers what it
needs, reconciles what already exists, and hands off.

**Two entry points, same flow:**
- **Fresh setup** — no `AGENTS.md` yet: interview, then write the standard.
- **Adopt / update** — a repo that already has agent docs (or a buckle-style hub):
  read the current state with `gyroscope check`, then converge it (see "Existing
  repo — reconcile, don't clobber"). This is the path for "get this repo up to date
  with gyroscope".

<HARD-GATE>
This skill is model-invocable, so it may fire from a plain request like "get this
repo up to date with gyroscope" — that convenience does NOT relax the gate. Do NOT
write any file, run `gyroscope init --apply`, or `--force` anything until you have
shown the plan (the classified file list + any conflicts + the answers you gathered)
and the user has approved it. Present first, even if the request seems fully
specified. `gyroscope check` and dry-run `gyroscope init` are read-only and may be
run before approval to build that plan.
</HARD-GATE>

## Explore first, then ask

Before asking anything, scan the repo: build files, test scripts, an existing
`CLAUDE.md`/`AGENTS.md`, the language/framework. Answer what the code can tell you;
only ask the user what it cannot.

## Interview — one question at a time

Ask relentlessly, ONE question per message, recommending an answer each time. Walk
each branch to the end before moving on. Asking several at once is bewildering.
Gather:

1. **Project one-liner** — what is this repo, in a sentence? _(fills `AGENTS.md`, `CONTEXT.md`)_
2. **Domain terms agents get wrong** — words used two ways, jargon to pin down. _(fills `CONTEXT.md`)_
3. **Build & test** — how to build; the test command that gates "done". _(fills `docs/agents.md`)_
4. **Conventions & do-nots** — key patterns and explicit no-gos, each with its reason. _(fills `docs/agents.md`)_
5. **Spokes** — accept the default standard, or add/drop any? _(sets the `gyroscope` config)_

Keep track of which answer fills which file — the fill step below needs that mapping.
Placeholders: `{{...}}` are yours to fill in step 3; `<...>` form fields in
`docs/adr/TEMPLATE.md` stay.

## Present, then install, then fill

1. Show the gathered answers and the exact list of files to be written. Get approval.
2. Run the binary to write the standard, reconcile pointer files to the one routing
   line, and install the `SessionStart` hook (merged into `.claude/settings.json`,
   never overwriting existing content). This step shells out to the `gyroscope`
   binary — if `gyroscope` is not on PATH, `install-skill` already warned you and
   printed the `go install` line; run that first, or the write step fails with a
   bare "command not found". `init --apply` is merge-safe: it creates missing
   files, injects the hub's managed region into an existing `AGENTS.md` without
   touching your other content, and only a genuine CONFLICT needs `--force`.
3. **Fill the scaffolds.** The binary writes them with `{{...}}` placeholders; you fill
   them from the approved answers. For each written spoke — `CONTEXT.md`,
   `docs/agents.md`, `docs/agents/README.md`, and the date in `docs/adr/0001-*` —
   replace every `{{...}}` with the corresponding answer and delete the scaffold-guidance
   `<!-- ... -->` comments. Do NOT invent content the interview didn't cover; drop a term
   or leave a section terse rather than padding.
   Leave `docs/adr/TEMPLATE.md` untouched — it is a blank form whose `<...>` fields are
   filled per-ADR, not now. The process artifacts (`CONTRIBUTING.md`,
   `.github/pull_request_template.md`, `.gitmessage`) are written ready-to-use with no
   `{{...}}` to fill; `CONTRIBUTING.md` intentionally defers conventions to
   `docs/agents.md` — don't copy them in.
4. **Verify none remain.** `grep -rn '{{' AGENTS.md CONTEXT.md docs/` must come back empty
   (every `{{...}}` is a fill-once placeholder; `<...>` in `TEMPLATE.md` is intentional and
   is not matched). If anything remains, you skipped a spoke — go fill it.
5. Confirm what landed and where.
6. **Optional — mention the complement.** gyroscope shapes what agents *read*. If the
   user also finds agent *replies* too verbose (the usual cause of rubber-stamping),
   point them at `caveman` (`github.com/JuliusBrussee/caveman`) — a complementary,
   output-side skill that makes replies terse while keeping code/commands byte-exact.
   It coexists with gyroscope's hook (both append-merge into `SessionStart`). Suggest,
   don't insist — it's a third-party tool, not part of the standard.

## Existing repo — reconcile, don't clobber

When the repo already has agent docs (an `AGENTS.md`, a `CLAUDE.md` with real
content, a buckle-style hub, `CONTRIBUTING.md`, etc.), do NOT run the fresh-setup
path blind and do NOT reach for `--force` — it overwrites content-bearing files.
Converge instead:

1. **See the current state (read-only).** `gyroscope check .` lists every drift;
   `gyroscope init .` (dry-run) classifies each destination **NEW** (create),
   **OK** (already conformant), **MERGE** (present, gyroscope injects its managed
   region and leaves the rest), or **CONFLICT** (present, differs — needs a
   decision). Build the plan from this.
2. **Present the plan, get approval** (the HARD-GATE). Show the classified list,
   name the CONFLICTs, and say what you intend to do with each.
3. **Apply the safe subset:** `gyroscope init --apply .` — **no `--force`**. This
   creates the NEW files, injects the hub's managed region into an existing
   `AGENTS.md` without touching the user's other content, and **skips** every
   CONFLICT (it prints them and exits with drift). Nothing is destroyed.
4. **Reconcile each CONFLICT by hand, one at a time** — never a blanket `--force`:
   - **A content-bearing pointer** (e.g. a `CLAUDE.md` that holds real instructions):
     gyroscope wants it reduced to the one routing line. Relocate its content into
     the hub or the matching spoke first, then replace the file with the canonical
     routing line.
   - **`CONTRIBUTING.md` / `.gitmessage` / `.github/pull_request_template.md` /
     `docs/agents/README.md`:** compare the repo's version against gyroscope's. If
     the repo's already satisfies the artifact, keep it (it's fine that it differs);
     adopt gyroscope's only if the user prefers it, and only then `--force` that one
     file.
   - **Hub routes:** after the merge the hub may carry both the repo's own routes and
     gyroscope's managed `## Routes` block. Fold the custom routes into the managed
     block, or leave them under a separate heading OUTSIDE the
     `<!-- gyroscope:managed -->` markers — content outside the markers is the user's
     and is invisible to `check`. Remove the duplication either way.
   - **Existing TODO state → split into `TODO.md` + `DONE.md`:** `TODO.md` is catted
     into every session, so it must stay lean. Move every completed `[x]` item out of
     the repo's TODO into `DONE.md` (the routed-but-not-injected archive), leaving only
     open `[ ]`/`[~]` work in `TODO.md`. If the repo has a stray or non-root todo (e.g.
     a `docs/TODO.md`, or an agent-invented "lean root + detailed" second file),
     consolidate it into the root `TODO.md` in the same pass — **never leave two todo
     files.** This mechanical status boundary replaces any prior lean/detailed
     convention (see ADR 0009).
5. **Fill placeholders from what already exists.** On an established repo, prefer the
   repo's own docs/READMEs over interviewing from scratch — pull the one-liner,
   domain terms, and build/test commands from what's there, and only ask the user
   for what the repo can't answer.
6. **Converge to green:** `gyroscope check --fix .` re-applies the safe convergence
   (never clobbers a CONFLICT) and re-checks; loop with hand-reconciliation until
   `gyroscope check .` reports `conformant` (exit 0).

## Agent personas (docs/agents/)

The hub carries a standing rule: when `gyroscope.json` `spokes.personas` is
`unknown`, ask the user about personas before other work. Also run this flow when
the user invokes `/gyroscope` explicitly.

**First — reconcile what's already on disk (do this before the wire/skip decision,
regardless of `spokes.personas` state).** `gyroscope check` is *config-relative*: it
verifies the repo against `gyroscope.json`, so a `skipped`/`unknown`/`off` state means
it never inspects persona files and a green `check` says **nothing** about them. Do not
read green as "personas are handled." Look yourself:

- Scan for persona-shaped files anywhere under `docs/agents/`, recursively — the mirror
  reads only the root (`docs/agents/*.md`), so a subdir is exactly where a persona hides
  unregistered: `grep -rl '^name:' docs/agents --include='*.md'`.
- Classify each hit:
  - **Root + `name:` frontmatter** (`docs/agents/<file>.md`) → canonical (ADR 0010).
    Registers into `.claude/agents/` on `init --apply` once state is `on`.
  - **In a subdir** (`docs/agents/personas/…`, etc.) → off the mirror's path, so it will
    **never** register as a subagent. It is one of two things, and you must **ask** which
    — do not assume:
    - (a) a **doc-routed** persona, dispatched by prompt-injection from a runbook (e.g.
      `review-checkpoints.md` references it). A valid, deliberate pattern — leave it.
    - (b) a persona the author meant to register but mislaid. Move it to
      `docs/agents/<name>.md`, kebab-case the `name:`, add a `description:`, then register.
- Reconcile state-vs-disk when they disagree:
  - Files present but state is `skipped`/`unknown` → surface it: "found N persona files
    but the config says `<state>` — register them (root + `agents set on` + `init
    --apply`), keep them doc-routed as-is, or leave skipped?" Let the user decide.
  - State `on` but no root `name:` file → the mirror registers nothing; the files are in
    a subdir or lack frontmatter. Fix the **layout**, don't just trust the flag.
- If the user keeps a doc-routed setup on purpose, record the decision (`agents set
  skipped` is fine) and stop — do not re-nag next session.

Only once the reconcile is settled, run the wire/skip decision:

1. **Ask:** "Wire agent personas for this repo now, or skip for now?"
2. **Skip** → run `gyroscope agents set skipped`. Stop; do not revisit unless asked.
3. **Wire** →
   a. Ask for the persona template directory. Offer `~/Src/agency-agents` as a
      guessed default, but do not persist the answer — ask again next time.
   b. List the persona files under that directory (they are grouped by division,
      e.g. `engineering/`, `testing/`). Present them and let the user pick a subset.
   c. For each pick, read the template and **customize it to this repo**: keep the
      useful framing, adapt language to this repo's stack, and strip content that
      does not apply (web-only stacks, framework-specific tooling). Write the
      result to `docs/agents/<name>.md` — one file per persona, in Claude subagent
      format: YAML frontmatter plus the system-prompt body. The frontmatter must
      carry:
      - `name:` **kebab-case** (`data-engineer`, not `Data Engineer`) — it becomes
        both the `.claude/agents/<name>.md` filename and the dispatchable subagent
        *type*, and a space breaks the type. Match the filename to it.
      - `description:` a one-line trigger — Claude's auto-dispatch matches the task
        against it, so a persona with no `description:` still registers but is
        effectively never auto-selected.
      - optional `tools:` / `model:`.
   d. Run `gyroscope agents set on`.
   e. Run `gyroscope init --apply` to **register** the personas: the binary mirrors
      each valid `docs/agents/` persona into every enabled harness's native subagent
      registry so it can be dispatched as a subagent (gated on personas `on` + that
      harness's adapter). `docs/agents/` stays canonical — edit personas there and
      re-run; never hand-edit a generated mirror. `gyroscope check` flags a missing
      or drifted mirror; `check --fix` re-mirrors. Which registries are written
      depends on the `enforce.*` toggles:
      - **Claude** (`enforce.claude`, default on) → `.claude/agents/<name>.md`,
        byte-for-byte (ADR 0010).
      - **OpenCode** (`enforce.opencode`) → `.opencode/agents/<name>.md` and
        **Codex** (`enforce.codex`) → `.codex/agents/<name>.toml`: re-encoded from the
        same canonical source into each tool's format (ADR 0014).
      - **PI** (`enforce.pi`) → `.pi/extensions/gyroscope-personas.ts`, a generated
        extension registering each persona as a dispatchable subagent tool (ADR
        0015). PI has no native registry, so this is runtime, not a file mirror.
      A persona's `tools:` allowlist carries through to each (OpenCode `permission`,
      PI `--tools`, Codex `sandbox_mode`) where the harness supports it.

Persona-fit guidance for a Go/CLI backend repo:
- code-reviewer templates → quality-review framing.
- reality-checker templates → spec-review mindset (default to NEEDS WORK; verify
  by reading code and running commands). Ignore any web/Playwright mechanics.
- minimal-change / backend-architect templates → implementer framing (YAGNI,
  minimal diff).
- Avoid generic "senior web developer" personas — they inject irrelevant web/UI
  concerns into a Go CLI.

The binary never **authors** persona content — that is entirely this skill's job.
The binary records the decision (`gyroscope agents set …`), scaffolds the empty
`docs/agents/` spoke, and **registers** personas into each enabled harness's
registry — a byte-copy for Claude (ADR 0010), a deterministic format re-encode for
OpenCode/Codex and the PI extension (ADR 0014/0015) — but it never writes the
content itself. Re-encoding the mirror files it owns is mechanical, not templating
of user content.

## Harness enforcement (opt-in per tool)

gyroscope installs enforcement per harness. Claude is on by default; the rest are
opt-in via `gyroscope.json` `enforce.*` (re-run `gyroscope init --apply` after
changing them, then `gyroscope check`). Each toggle does what that harness's
mechanism actually allows — verified per tool, not assumed:

- **`claude`** (default `true`) — merges a `SessionStart` hook into
  `.claude/settings.json` that cats the hub + key spokes each session, and mirrors
  personas into `.claude/agents/`.
- **`pi`** — writes `.pi/extensions/gyroscope-context.ts` (a `session_start`
  extension injecting the non-hub state spokes; PI reads `AGENTS.md` natively) and,
  with personas on, `.pi/extensions/gyroscope-personas.ts` (persona subagents).
  Then run `/trust` once in PI for this project — it loads the extensions only after
  the project is trusted. gyroscope does **not** write PI's trust decision.
- **`opencode`** — reads `AGENTS.md` natively, so enabling it wires the parts that
  aren't free: persona registration into `.opencode/agents/` (see personas above), and
  it merges the non-hub **state spokes** into `opencode.json` `instructions` (OpenCode's
  always-load mechanism — the analog of Claude's hook, since native read covers only
  the hub, not the resume state). `.local/*` personal spokes are excluded (gitignored;
  no missing-file tolerance).
- **`codex`** — reads `AGENTS.md` natively; persona registration into `.codex/agents/`
  (see personas above), plus `commit_attribution` when attribution is suppressed (below).
- **`cursor`** — writes `.cursor/rules/gyroscope-hub.mdc` with `alwaysApply: true`,
  a rule force-injected into every Cursor request ("read AGENTS.md and follow its
  routes") — the enforcement analog above the passive `.cursorrules` pointer (ADR
  0016). **Zed needs no toggle**: it auto-reads the `.rules` pointer every session.
- **`aiAttribution`** (default `true`) — set `false` to suppress AI-authorship
  credit repo-wide: writes Claude's `includeCoAuthoredBy:false` and Codex's
  `commit_attribution=""` (when those adapters are on) and injects a cross-harness
  hub directive so every harness honors it. Scoped to AI credit, not the developer's
  own DCO `Signed-off-by`.

## Cross-harness slash commands (`docs/commands/`)

Author a slash command ONCE in `docs/commands/<name>.md` and gyroscope re-encodes it
into every enabled harness (ADR 0019): `.claude/commands/<name>.md`,
`.opencode/commands/<name>.md`, `.pi/prompts/<name>.md` — dispatchable as `/name`
everywhere. Write the canonical file in the **Claude command format** (the superset):
optional YAML frontmatter (`description`, `argument-hint`, `allowed-tools`, `model`)
plus the prompt body. The body syntax is portable — `$ARGUMENTS`/`$1`, `` !`cmd` ``
shell injection, `@file` references all work in every target — so only the frontmatter
is reduced per harness (OpenCode keeps `description`; PI keeps `description` +
`argument-hint`; Claude-only keys are dropped, not mistranslated).

- **Gating:** per enabled harness (`enforce.{claude,opencode,pi}`), presence-driven —
  a repo with `docs/commands/*.md` gets them; no separate toggle, and **independent of
  the personas state**.
- **Adopt existing:** if the repo already has `.claude/commands/*.md`, move them into
  `docs/commands/` once (like persona adoption) so the other harnesses get them too;
  then re-run `gyroscope init --apply`.
- Generated command files are gyroscope-owned mirrors — `check` verifies them,
  `check --fix` re-encodes, and `gyroscope-sync.sh` (if `sync` is on) bakes them for a
  binary-less clone. The binary re-encodes; it never authors the command body.

## MCP servers (`gyroscope.json` `mcp`)

Declare shared MCP servers once in `gyroscope.json` and gyroscope re-encodes them per
enabled harness (ADR 0020):

```json
{ "mcp": {
    "linear": { "command": "npx", "args": ["-y","linear-mcp"], "env": { "KEY": "${LINEAR_KEY}" } },
    "docs":   { "url": "https://x/mcp", "headers": { "Authorization": "${DOCS_TOKEN}" } }
} }
```

A server is stdio/local when it has a `command`, remote when it has a `url`.

- **Secrets:** use `${ENV_VAR}` references only, never raw keys — every target expands
  env vars, so the generated configs are safe to commit and to bake into the sync
  script. gyroscope passes the refs through verbatim.
- **Targets:** Claude → dedicated `.mcp.json` (gyroscope-owned whole file, gitignorable,
  reconstructed by sync). OpenCode → merged into `opencode.json` `mcp` (shared config;
  gyroscope's declared servers are added by name, the user's other entries preserved —
  but a server removed from `gyroscope.json` is not auto-pruned from the shared file).
- **Gating:** `mcp` non-empty AND `enforce.<harness>`. Absent `mcp` = feature off.
- **PI** (`enforce.pi`) → gyroscope generates `.pi/mcp.json` PLUS a minimal MCP-client
  extension `.pi/extensions/gyroscope-mcp/` (`index.ts` + `package.json`) since PI has
  no native MCP — using only the OFFICIAL `@modelcontextprotocol/sdk` (no community
  wrapper; ADR 0021). One-time teammate step: `npm install` in that dir. Stdio servers
  only for now (remote `url` written to config but not yet connected). Validated live.

## Code-review preset (opt-in)

An alternative to the from-scratch persona interview: a curated, multi-persona
**code-review** bundle. Trigger it when `gyroscope.json` has `"codeReview": true`, or
when the user asks for "a review suite / reviewer personas / code-review setup." The
binary only records the flag (`{ "codeReview": true }`); this skill authors the
bundle — the personas are customized per repo exactly like normal personas (register
them through the persona flow above). See ADR 0017.

**Scaffold these into `docs/agents/` (repo-tuned — include only the specialists a
reviewer would actually apply to THIS repo; drop genres that don't exist here):**

- **Layer-aligned reviewer personas** — one `*-reviewer.md` per relevant area, each
  with a `description:` trigger and an `appliesTo:` glob list scoping it to changed
  files. The canonical menu (from a web app; keep what fits): `backend-reviewer`
  (logic/conventions/scale), `frontend-reviewer`, `ux-reviewer`, `experience-reviewer`
  (whole-feature product fit), `design-reviewer` (visual/brand), `db-reviewer`
  (schema/queries/indexes/N+1), `cicd-reviewer` (pipeline/containers/deploy),
  `security-reviewer` (auth/input/secrets/injection/authz). For a Go CLI, that's
  typically **backend + security + cicd** only — customize each to the repo's stack,
  strip web/UI framing.
- **`security-reviewer`** — always runs (security issues appear in any diff).
- **`synthesizer.md`** — aggregates the specialists' findings and **never
  pre-filters**: consensus (2+), unique-per-specialist, cross-cutting, priorities.
- **`intent-verifier.md`** — runs **last**, grounding each finding against the repo's
  documented intent (`docs/adr/`, `TODO.md`, `DONE.md`, the spokes): UPHELD /
  OVERRULED-with-citation / NEEDS-DECISION, and independently scans the diff for
  intent-violations the reviewers missed. This is the load-bearing piece — it catches
  the "technically right but wrong for this repo" false positive and the silent
  regression against a prior decision. Distinct from `reality-checker` (which *runs*
  checks to prove a claim), so keep both if the repo has one.
- **`code-review.md` orchestrator runbook** — the workflow: identify the diff, select
  specialists by `appliesTo` match (security always on), dispatch in parallel,
  synthesize, then intent-verify, then present **both** blockers and non-blocking
  findings (a clean verdict means "no blockers," not "no findings").
- **The grounding ADR** — record an ADR in the repo's `docs/adr/` ("Ground code
  review against documented intent") capturing *why* the intent-verifier exists, so
  the decision is on the record like any other. It complements Claude's built-in
  `/code-review` (that is a generic runner; this is the repo's own reviewers +
  intent-grounding).

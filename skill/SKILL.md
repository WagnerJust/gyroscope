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
      result to `docs/agents/<name>.md` — one file per persona.
   d. Run `gyroscope agents set on`.

Persona-fit guidance for a Go/CLI backend repo:
- code-reviewer templates → quality-review framing.
- reality-checker templates → spec-review mindset (default to NEEDS WORK; verify
  by reading code and running commands). Ignore any web/Playwright mechanics.
- minimal-change / backend-architect templates → implementer framing (YAGNI,
  minimal diff).
- Avoid generic "senior web developer" personas — they inject irrelevant web/UI
  concerns into a Go CLI.

The binary never reads or writes persona content — that is entirely this skill's
job. The binary only records the decision (`gyroscope agents set …`) and scaffolds
the empty `docs/agents/` spoke.

## PI enforcement (opt-in)

gyroscope installs enforcement per harness. Claude is on by default. To enforce
the hub for the **PI** coding agent, enable it in `gyroscope.json`:

```json
{ "enforce": { "pi": true } }
```

Re-run `gyroscope init` to write `.pi/extensions/gyroscope-context.ts`. Then, in
PI, run `/trust` once for this project — PI reads `AGENTS.md` and loads the
extension only after the project is trusted. gyroscope does **not** write PI's
trust decision; that is your security choice.

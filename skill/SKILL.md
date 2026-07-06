---
name: gyroscope
description: Set up gyroscope's opinionated, self-enforcing agent-doc standard in this repo — interview, then write the standard and install the enforcement hook.
disable-model-invocation: true
---

# gyroscope

Interview the user, then install gyroscope's standard: an opinionated hub-and-spoke
doc set (`AGENTS.md` hub, `CONTEXT.md` glossary, `docs/agents.md` instructions,
`docs/adr/`, `docs/agents/` personas), the L2 process artifacts (`CONTRIBUTING.md`
routed from the hub, plus `.github/pull_request_template.md` and `.gitmessage`,
which Git/GitHub apply directly), and a `SessionStart` hook that injects the hub +
instructions every session. The `gyroscope` binary does the writing — this skill
gathers what it needs and hands off.

<HARD-GATE>
Do NOT write any file or run `gyroscope init` until the user approves the gathered
answers. Present them first, even if the request seems fully specified.
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
   never overwriting existing content).
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

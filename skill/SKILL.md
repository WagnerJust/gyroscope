---
name: gyroscope
description: Set up gyroscope's opinionated, self-enforcing agent-doc standard in this repo — interview, then write the standard and install the enforcement hook.
disable-model-invocation: true
---

# gyroscope

Interview the user, then install gyroscope's standard: an opinionated hub-and-spoke
doc set (`AGENTS.md` hub, `CONTEXT.md` glossary, `docs/agents.md` instructions,
`docs/adr/`, `docs/agents/` personas) plus a `SessionStart` hook that injects the
hub + instructions every session. The `gyroscope` binary does the writing — this
skill gathers what it needs and hands off.

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

1. **Project one-liner** — what is this repo, in a sentence? _(seeds `AGENTS.md`, `CONTEXT.md`)_
2. **Domain terms agents get wrong** — words used two ways, jargon to pin down. _(seeds `CONTEXT.md`)_
3. **Build & test** — how to build; the test command that gates "done". _(seeds `docs/agents.md`)_
4. **Conventions & do-nots** — key patterns and explicit no-gos, each with its reason. _(seeds `docs/agents.md`)_
5. **Spokes** — accept the default standard, or add/drop any? _(seeds the `gyroscope` config)_

## Present, then install

1. Show the gathered answers and the exact list of files to be written. Get approval.
2. Run the binary to write the standard, reconcile pointer files to the one routing
   line, and install the `SessionStart` hook (merged into `.claude/settings.json`,
   never overwriting existing content).
3. Confirm what landed and where.

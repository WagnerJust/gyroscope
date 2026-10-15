# ADR 0022: Rules reach Claude by import, state by a budgeted hook

- **Status:** accepted
- **Date:** 2026-10-14

## Context

Since the MVP the Claude adapter has delivered everything through one channel: a
`SessionStart` hook that `cat`s the hub, `docs/agents.md`, the state files and the
personal `.local/` files into context. ADR 0005, ADR 0011 and ADR 0018 each chose
to keep that hook a pure `cat` — no binary dependency, nothing that writes — and
the standard's premise rested on it: the rules and the open work are in context
from the first turn, so following the docs and resuming work "isn't left to
chance."

Two facts, both checked on 2026-10-14, show the premise did not hold:

- **Claude Code caps a hook's stdout (and its `additionalContext`) at 10,000
  characters, with no setting to raise it.** Past the cap it saves the output to a
  file and injects a 2,000-character preview; it does not ask Claude to read the
  file (code.claude.com/docs/en/hooks). The standard's base payload — hub,
  `docs/agents.md`, the `.local/` files and `gyroscope.json`, before `TODO.md`
  adds a byte — is 5.6–12 KB in the repos that run it, and `TODO.md` alone
  crossed 10 KB within two days of adoption in carsearchhub, one in prony, three
  in this repo. Session transcripts put the failure at every actively used repo:
  this repo 9 of 9 sessions truncated, carsearchhub 118 of 120, prony 90 of 92,
  the wiki 32 of 33. What arrived was the first 2 KB of `AGENTS.md`. `check`
  reported `conformant` throughout, because it verified the configured command
  string and never looked at what the command produced.
- **The harness has a separate, uncapped channel built for exactly the static
  half of that payload.** `CLAUDE.md` and the files it imports with `@path` load
  in full at launch (up to 4 MiB per file), are re-injected after compaction, and
  the docs recommend them for "static context that doesn't require a script"
  (code.claude.com/docs/en/memory). Block-level HTML comments are stripped at
  load, so the managed markers cost no context there. The same page names
  gyroscope's two mechanisms as workarounds to remove: a `CLAUDE.md` that tells
  Claude *in words* to read `AGENTS.md` ("Claude sees AGENTS.md only if it decides
  to open the file" — replace the sentence with an `@AGENTS.md` import), and a
  `SessionStart` hook that prints `AGENTS.md` (a second copy once the import loads
  it). Claude Code 2.1.277+ also reads `AGENTS.md` natively, but only when no
  `CLAUDE.md` or `CLAUDE.local.md` exists in or above the working directory.

The genuine alternatives:

- **Delete `CLAUDE.md` and rely on native `AGENTS.md` reading.** Rejected: the
  native read is version-gated, and any `CLAUDE.local.md` a developer adds
  silently switches it off. The import is the form the harness documents as
  durable, and its docs say keeping it "never makes Claude read AGENTS.md twice."
- **A `gyroscope session-context` subcommand that computes the injection.**
  Rejected again, for ADR 0005's reason: a teammate who never installed gyroscope
  must still get the standard, so the hook may not depend on the binary.
- **Raise the cap.** Not possible — and the evidence on long-context degradation
  argues against wanting a 300 KB payload in context even if it were.
- **Keep one channel and shrink the payload under the cap.** Rejected: the hub
  plus the personal files already sit near 10 KB in an ordinary repo, which leaves
  no room for the state the `state` spoke exists to carry.

## Decision

Split the Claude adapter's delivery into the two channels the harness provides,
by the kind of content:

- **Static rules go through the memory channel.** The Claude pointer file is no
  longer the routing sentence; it is an import pointer whose body is exactly

      <!-- gyroscope: this file imports the hub so Claude Code loads it in full at launch. Edit AGENTS.md, not this file. -->
      @AGENTS.md
      @docs/agents.md

  with the third line present only when the `agents` spoke is on. The other six
  pointer files keep the single routing line — their tools have no import syntax
  gyroscope has verified. A `CLAUDE.md` whose bytes equal the old routing line is
  gyroscope's own output, so `init --apply` and `check --fix` upgrade it in
  place; any other content is still a CONFLICT for the skill's reconcile flow.
- **Dynamic state goes through the hook, within a budget.** The single hook path
  list becomes two — rule paths (`AGENTS.md`, `docs/agents.md`) and state paths
  (`gyroscope.json`, `TODO.md`, `.local/todo.md`, `.local/local.md`, in that
  priority order) — and the Claude hook carries only the state. The command is a
  POSIX `sh` loop rather than a `cat`: it prints a one-line preamble saying the
  files below are working state and that rules live in `AGENTS.md`, then for
  each state file in order either injects it under a `### <path> (<bytes> bytes)`
  header when it fits the remaining 9,000-byte budget, or prints a one-line
  pointer saying it is over budget and not injected. Bytes are counted with
  `wc -c`, so multibyte text is counted conservatively against the character
  cap. The command ends with a `# gyroscope-hook` tag, and `Install` replaces an
  entry carrying the tag — or the legacy `cat AGENTS.md …` prefix — instead of
  appending beside it. PI and OpenCode keep injecting what they did; they only
  receive the two lists.
- **`check` measures delivery, not just installation.** It simulates the hook's
  budget over the files present: a `TODO.md` larger than the budget is **drift**,
  because it can never be injected; any other deferred state file is a soft note.
  An `AGENTS.md` or `docs/agents.md` over 200 lines is a soft note quoting the
  harness's own adherence guideline, since the import channel now delivers those
  files whole.

The pure-`cat` property is given up. The property it protected is kept: the hook
depends on nothing but `sh`, it only reads and prints, and nothing it runs writes.

## Consequences

- In a repo under the budget nothing visible changes, except that the hub and
  conventions now arrive through `CLAUDE.md`'s imports and survive compaction. In
  a repo over it, the hub and conventions still arrive whole, the state degrades
  file by file with a pointer line per casualty, and `check` says so — the failure
  stops being silent at both ends.
- `TODO.md` acquires a hard size rule for the first time: one line per open item,
  narrative in a `DONE.md` entry or an ADR, and under 9,000 bytes or `check`
  reports drift. `check --fix` cannot trim it (that is judgment), so the drift
  persists until a person or the skill's reconcile step resolves it — the same
  shape as an unfilled placeholder.
- The hub prose and the contributor block change ("CLAUDE.md imports the hub; the
  rest hold one line"), so every adopter's managed regions re-converge on the
  next `check --fix`, as do their `CLAUDE.md` and hook entry.
- Two implementations of the budget exist — the generated `sh` loop and the Go
  simulation `check` runs — the dual-implementation trap ADR 0018 named. The
  mitigation is an end-to-end test that executes the generated command against
  files of known sizes and asserts its output matches the Go plan.
- The import channel delivers a large `docs/agents.md` in full, where the cap used
  to drop it. A repo whose conventions file runs to tens of kilobytes gets worse,
  not better, until that file is trimmed or split; the 200-line note exists to
  say so before the pointer is switched.
- Delivery is now checkable from inside a session: `/context` lists `CLAUDE.md`
  and its imports under Memory files, and the hook's output shows which state
  files landed. The skill's confirm step says to look.

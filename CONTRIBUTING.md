# Contributing

This is the canonical guide for how changes get proposed and reviewed in this
repo — for human contributors and AI agents alike. It covers the *process*.
The build/test commands, conventions, and do-nots live in `docs/agents.md`;
this guide defers to that spoke rather than repeating it, so the two never
drift.

## Workflow

1. Branch from the main branch: `git checkout -b <type>/<short-name>`.
2. Make focused commits — one logical change per commit, one concern per PR.
3. Run the test suite before opening the PR. See `docs/agents.md` for the exact
   build and test commands that gate "done".
4. Open a PR and fill in the checklist from the pull-request template.

## What makes a change mergeable

- It does one thing; unrelated cleanups go in their own PR.
- Tests cover the new behavior and the whole suite passes locally.
- It follows the conventions and avoids the do-nots in `docs/agents.md`.
- The diff has no stray formatting churn in untouched lines.

## Review

- Reviewers read for correctness, scope, and fit with the conventions above.
- Respond to review on the branch; don't force-push over discussion history
  unless asked.
- Raise a stuck branch early — open a draft PR or an issue rather than sitting
  on it.

## Commit messages

Follow the repo's commit convention (see `.gitmessage`): a concise
`<type>: <imperative summary>` subject, with a body explaining *why* when the
diff isn't self-explanatory.

<!-- gyroscope:managed -->
## Working with the gyroscope-maintained docs

This repo's agent-facing docs — the `AGENTS.md` hub, the spokes it routes to, the
pointer files (`CLAUDE.md`, `.cursorrules`, …), and any `.claude/agents/` personas
— are scaffolded and kept in sync by [gyroscope](https://github.com/WagnerJust/gyroscope).
**You do not need gyroscope installed to read these docs or to contribute.** They
are plain Markdown; read and edit them like any other file, and open PRs the normal
way described above.

A few gyroscope-specific things you may notice, and what they mean:

- Paired `gyroscope:managed` / `/gyroscope` HTML-comment markers fence the regions
  gyroscope owns (including this section). You *may* edit inside a fenced region and
  nothing breaks — but the next `gyroscope check --fix` a maintainer or CI runs
  reconciles that region back to the standard, so it may overwrite your change. Put
  durable prose **outside** the fences, where it is yours to keep.
- The root pointer files (`CLAUDE.md`, `GEMINI.md`, `.cursorrules`, `.windsurfrules`,
  `.clinerules`, `.rules`, `.github/copilot-instructions.md`) each hold a single line
  pointing their AI tool at `AGENTS.md`. They are redirects, not config to edit — the
  real content lives in the hub and its spokes.
- `docs/agents.md` (a file) is the build/test/conventions spoke; `docs/agents/` (a
  directory) holds agent personas. Similar names, different things.
- `gyroscope.json` records which doc spokes are enabled. Harmless to ignore.
- `.gitmessage` is a commit-message template; it only takes effect once you run
  `git config commit.template .gitmessage` (nothing wires it automatically).
- `.local/` is gitignored personal scratch — never committed, never shared.

You never have to run gyroscope — these docs are plain Markdown and hand-edits are
safe. If a maintainer or CI happens to run `gyroscope check` it reconciles any
drift; if no one does, nothing breaks. If you *want* to check or converge them
yourself without installing anything, it is a single dependency-light Go binary,
runnable straight from source:

    go run github.com/WagnerJust/gyroscope/cmd/gyroscope@latest check .

Add `--fix` to converge the safe drift. That is entirely optional.
<!-- /gyroscope -->

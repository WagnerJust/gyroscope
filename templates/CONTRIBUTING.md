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

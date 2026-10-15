# Agent instructions — {{PROJECT}}

Applies to all work in this repo.

## Build & test

- {{How to build.}}
- Run `{{test command}}` before claiming a change is done.

## Conventions

- {{Language / framework and the version pin.}}
- {{Key patterns to follow — with the reason, not just the rule.}}
- State hygiene: `TODO.md` holds open work only, one line per item, and is
  injected every session while it fits the 9,000-byte session budget; when a task
  is done, move its line to `DONE.md` (the archive is routed, not injected), and
  put narrative there or in an ADR rather than in the open list, so the
  per-session context stays small.

## Do NOT

- {{Explicit no-go, and *why* it bites — e.g. "NEVER commit generated files;
  they drift from source and mask real diffs."}}

<!-- Imperative voice. State negative constraints with their reason. Backticks
for paths/commands/versions. gyroscope seeds this from the interview + a scan of
the repo (build files, test scripts, existing config). -->

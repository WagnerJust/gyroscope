# ADR 0018: sync script bakes rendered bytes, idempotent, opt-in

- **Status:** accepted
- **Date:** 2026-07-24

## Context

gyroscope's "ships to the team" promise breaks when a repo gitignores the generated
dirs (`.claude/`, `.opencode/`, `.codex/`, `.pi/`): a teammate who clones and never
installed gyroscope gets the tracked `docs/agents/` personas but none of the
generated subagent *registrations*, so nothing is dispatchable. The already-shipped
ship-to-team note offers one fix — commit the generated dirs via a `!`-negation. This
ADR is the second, for teams that will not commit generated files: reconstruct them
per-dev from a committed, dependency-free script.

Two forces shaped the design:

- **bash cannot reproduce the mirrors.** The original sketch was a `cp` copy-list
  (`docs/agents/x.md` → `.claude/agents/x.md`). That only works for Claude, whose
  mirror is a byte-copy. The OpenCode/Codex/PI mirrors are *format re-encodes* (ADR
  0014/0015) — a bash re-implementation would be a second encoder that drifts from
  the Go one (the dual-implementation trap).
- **A committed, auto-run writing script is a trust-surface change.** Making the
  pure-`cat` SessionStart hook *write* files, or auto-running a committed script,
  expands what runs on the user's machine without an explicit choice.

## Decision

Generate `gyroscope-sync.sh` when a `gyroscope.json` `sync` flag (default false) is
set. It is:

- **Baked bytes, not a copy-list.** gyroscope renders every persona-registration
  mirror (`planPersonaMirrors`, all enabled harnesses) and bakes each file's exact
  bytes into the script as base64, with a tiny `w` helper that base64-decodes to the
  destination. base64 sidesteps shell-escaping arbitrary mirror content and lets one
  generator handle every harness format — no bash-side parsing or re-encoding. The
  script is regenerated from the mirror plan on `init --apply` / `check --fix`, and
  `check` verifies it byte-for-byte, so it never drifts from the encoders.
- **Idempotent (skip-if-exists).** `w` writes a file only if absent, so running the
  script on a fresh clone reconstructs the missing registrations and running it again
  is a no-op — it "syncs if it hasn't been," and never clobbers a dev's own agent or
  a committed mirror.
- **Opt-in, and dev-run — not wired into the hook.** The `sync` toggle is the user's
  explicit consent that this script exists; it defaults off. The script is a normal
  committed file a teammate runs after clone (or wires into their own tooling). It is
  deliberately NOT auto-invoked from the SessionStart hook: keeping the hook pure-cat
  avoids turning a read-only injection into a writer and avoids auto-running a
  committed script. (Auto-invocation from the Claude/PI hook is a possible future
  add-on behind the same toggle, if wanted.)

## Consequences

- The ship-to-team gap has two supported answers: commit the generated dirs (simple,
  advised by the existing note), or set `sync` and distribute the reconstruction
  script (for teams that keep generated files out of git).
- The script stays correct by construction — it carries the encoders' *output*, not a
  reimplementation — and `check` guards it, so a persona change that isn't followed by
  `check --fix` surfaces as drift rather than a stale script.
- POSIX sh + base64 only; Windows is a documented gap (as the SessionStart hook
  already assumes a shell).
- Scope is the persona-registration mirrors. The Claude SessionStart hook itself
  (`.claude/settings.json`) is not reconstructed by the script — a JSON merge is not
  safe to bake — so a Claude-only team that gitignores `.claude/` should still commit
  `settings.json` (the ship-to-team note already advises exactly that negation).

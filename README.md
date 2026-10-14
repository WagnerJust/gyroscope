# gyroscope

Install an opinionated, self-enforcing agent-doc standard into any repo:
an `AGENTS.md` hub routing to spokes (`CONTEXT.md`, `docs/agents.md`,
`CONTRIBUTING.md`, `docs/adr/`, `docs/agents/`, a repo-wide `TODO.md`, and
gitignored personal `.local/local.md` + `.local/todo.md`), the process artifacts
Git/GitHub enforce (`.github/pull_request_template.md`, `.gitmessage`), and two
Claude delivery channels: a `CLAUDE.md` that imports the hub and conventions so
Claude Code loads them in full at launch, and a budgeted `SessionStart` hook that
injects the current state — so following the docs, and resuming work, isn't left
to chance.

## Install
    go install github.com/WagnerJust/gyroscope/cmd/gyroscope@latest

## Use
    gyroscope install-skill --apply      # put the /gyroscope skill in ~/.claude/skills
    # then, in your agent:  /gyroscope   # interview → writes the standard
    # or non-interactively:
    gyroscope init            # dry-run: classify each file NEW/OK/MERGE/CONFLICT
    gyroscope init --apply    # write the standard + pointer + hook (merge-safe, idempotent)
    gyroscope check           # read-only: verify a repo still conforms (0=ok, 1=drift, 2=error)
    gyroscope check --fix     # auto-apply the safe convergence, then re-check

## Configure (optional)

Every spoke is on by default. To disable one, add `gyroscope.json` at the repo root:

    {"spokes": {"adr": false, "personas": false}}

Spokes: `context` (CONTEXT.md), `agents` (docs/agents.md), `contributing`
(CONTRIBUTING.md), `adr` (docs/adr/), `personas` (docs/agents/), `state`
(TODO.md + gitignored .local/todo.md, injected by the hook), `local`
(.local/local.md), `prTemplate` (.github/pull_request_template.md),
`commitConvention` (.gitmessage). The `AGENTS.md` hub is always written; disabling
a spoke also prunes its hub route.

### Enforcement adapters

`enforce` selects which harnesses are force-fed the hub. `claude` is on by default:
`CLAUDE.md` becomes an import pointer (`@AGENTS.md`, plus `@docs/agents.md`) so the
rules load in full through Claude Code's memory channel, and a `SessionStart` hook
in `.claude/settings.json` injects the state files (`gyroscope.json`, `TODO.md`,
`.local/todo.md`, `.local/local.md`) in that order while they fit a 9,000-byte
budget. Claude Code truncates hook output past 10,000 characters to a 2 KB
preview, so anything over budget is skipped with a one-line pointer instead, and
`gyroscope check` reports a `TODO.md` that can never fit as drift. `pi` (the Pi
Agent Harness) is opt-in — enable it and re-run init:

    {"enforce": {"pi": true}}

PI enforcement writes `.pi/extensions/gyroscope-context.ts`, which injects the
non-hub spokes on session start (PI reads `AGENTS.md` natively, so the hub is not
re-injected). PI loads the extension and reads `AGENTS.md` only after you `/trust`
the project in PI — gyroscope never writes PI's trust file.

`gyroscope init --apply` is merge-safe and idempotent: it classifies each
destination NEW / OK / MERGE / CONFLICT, creates the NEW files, injects the hub's
managed region (`<!-- gyroscope:managed -->` … `<!-- /gyroscope -->`) into an
existing hub while leaving your surrounding content untouched, and skips anything
already current. Only a genuine CONFLICT — a whole file that differs with no
managed region to merge into — needs `--force`. `gyroscope check --fix` applies the
same safe convergence, then re-checks; unresolved conflicts still report as drift.

## License

MIT — see [LICENSE](LICENSE).

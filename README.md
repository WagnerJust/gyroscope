# gyroscope

Install an opinionated, self-enforcing agent-doc standard into any repo:
an `AGENTS.md` hub routing to spokes (`CONTEXT.md`, `docs/agents.md`,
`CONTRIBUTING.md`, `docs/adr/`, `docs/agents/`, a repo-wide `TODO.md`, and
gitignored personal `.local/local.md` + `.local/todo.md`), the process artifacts
Git/GitHub enforce (`.github/pull_request_template.md`, `.gitmessage`), and a
Claude `SessionStart` hook that injects the hub + current state every session —
so following the docs, and resuming work, isn't left to chance.

## Install
    go install github.com/WagnerJust/gyroscope/cmd/gyroscope@latest

## Use
    gyroscope install-skill --apply      # put the /gyroscope skill in ~/.claude/skills
    # then, in your agent:  /gyroscope   # interview → writes the standard
    # or non-interactively:
    gyroscope init            # dry-run: show what would be written
    gyroscope init --apply    # write the standard + pointer + hook
    gyroscope check           # read-only: verify a repo still conforms (0=ok, 1=drift, 2=error)

## Configure (optional)

Every spoke is on by default. To disable one, add `gyroscope.json` at the repo root:

    {"spokes": {"adr": false, "personas": false}}

Spokes: `context` (CONTEXT.md), `agents` (docs/agents.md), `contributing`
(CONTRIBUTING.md), `adr` (docs/adr/), `personas` (docs/agents/), `state`
(TODO.md + gitignored .local/todo.md, injected by the hook), `local`
(.local/local.md), `prTemplate` (.github/pull_request_template.md),
`commitConvention` (.gitmessage). The `AGENTS.md` hub is always written; disabling
a spoke also prunes its hub route.

`gyroscope init --apply` refuses to overwrite existing files; pass `--force` to overwrite.

# gyroscope

Install an opinionated, self-enforcing agent-doc standard into any repo:
an `AGENTS.md` hub routing to spokes (`CONTEXT.md`, `docs/agents.md`, `docs/adr/`,
`docs/agents/`, and a gitignored personal `.local/local.md`), plus a Claude
`SessionStart` hook that injects the rules every session — so following the docs
isn't left to chance.

## Install
    go install github.com/WagnerJust/gyroscope/cmd/gyroscope@latest

## Use
    gyroscope install-skill --apply      # put the /gyroscope skill in ~/.claude/skills
    # then, in your agent:  /gyroscope   # interview → writes the standard
    # or non-interactively:
    gyroscope init            # dry-run: show what would be written
    gyroscope init --apply    # write the standard + pointer + hook

## Configure (optional)

Every spoke is on by default. To disable one, add `gyroscope.json` at the repo root:

    {"spokes": {"adr": false, "personas": false}}

Spokes: `context` (CONTEXT.md), `agents` (docs/agents.md), `adr` (docs/adr/),
`personas` (docs/agents/), `local` (.local/local.md). The `AGENTS.md` hub is always written.

`gyroscope init --apply` refuses to overwrite existing files; pass `--force` to overwrite.

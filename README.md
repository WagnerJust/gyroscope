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

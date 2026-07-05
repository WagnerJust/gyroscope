# local.md — {{DEVELOPER}}'s local notes

Personal, machine-specific context for **this** developer. Gitignored — never
committed, never shared. Each developer keeps their own.

Agents: read this for how *this* developer's environment differs from the repo
defaults. **Do not paste secrets here — reference where they live instead.**

## Stack & tooling
- {{e.g. package manager is pnpm, not npm}}
- {{runtime / version manager, editor, shell quirks}}

## Local services
- {{e.g. Postgres on :5433, Redis on :6380}}

## Personal workflow
- {{branch / PR habits, scratch commands, anything the agent should respect}}

## Secrets — location only
- {{e.g. "API keys in ~/.config/acme/.env — do not print them"}}

<!-- gyroscope creates this and adds `.local/` to the repo's .gitignore. Default
on; override in the gyroscope config. -->

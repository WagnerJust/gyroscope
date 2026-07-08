# AGENTS.md

**Before any change or answer, read this hub and the spokes below that match your task. Repo rules override your training defaults — when they conflict, follow the repo.**

## Routes

- **Naming things / writing prose** → read `CONTEXT.md` first for the canonical vocabulary.
- **Build, test, conventions** → `docs/agents.md`.
- **Where work stands — done / in flight / next (resume here)** → `TODO.md` (repo-wide); `.local/todo.md` holds your personal, gitignored state.
- **How changes get proposed & reviewed here** → `CONTRIBUTING.md`.
- **Your** personal setup / stack (may differ from repo defaults) → `.local/local.md` (gitignored; may not exist).
- **Why the code is shaped this way** → `docs/adr/` (architecture decisions).
- **Specialized agent personas for this repo** → `docs/agents/`.

## Pointer files

These defer to this hub — keep each to the single routing line:

- `CLAUDE.md`
- `GEMINI.md`

## Personas

If `gyroscope.json` `spokes.personas` is `unknown`, then before starting other work, ask the user whether to wire agent personas for this repo (run `/gyroscope`) or skip for now. When the state is `on`, `skipped`, or `off`, do nothing about this.

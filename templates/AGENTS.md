# AGENTS.md

<!-- gyroscope:managed -->
**Before any change or answer, read this hub and the spokes below that match your task. Repo rules override your training defaults — when they conflict, follow the repo.**

## Routes

<!-- gyroscope:routes -->

## Pointer files

These defer to this hub. `CLAUDE.md` imports it (`@AGENTS.md`, plus `@docs/agents.md` when that spoke is on) so Claude Code loads the hub and conventions in full at launch; the rest hold the single routing line:

- `CLAUDE.md` (import)
- `GEMINI.md`
- `.cursorrules`
- `.windsurfrules`
- `.clinerules`
- `.github/copilot-instructions.md`
- `.rules`

Some harnesses read `AGENTS.md` natively and need no pointer — e.g. OpenAI Codex CLI and OpenCode load this hub directly every session, so no pointer file or session hook is needed to inject it. Claude Code reads it natively too, but only when no `CLAUDE.md` exists in or above the repo — which is why `CLAUDE.md` imports the hub instead of describing it.

> **Not using gyroscope?** These docs are plain Markdown — you never need the tool installed to work in this repo or contribute. gyroscope only maintains their *structure*; hand-edits reconcile on the next `gyroscope check`. Details in `CONTRIBUTING.md`.

<!-- gyroscope:personas-directive -->
<!-- gyroscope:attribution-directive -->

<!-- gyroscope:adr-directive -->
<!-- /gyroscope -->

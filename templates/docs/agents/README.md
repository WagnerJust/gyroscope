# Repo agent personas

Specialized agent personas scoped to this repo. `AGENTS.md` routes here.

Add one markdown file per persona (e.g. `backend-reviewer.md`,
`migration-writer.md`) in Claude subagent format (YAML frontmatter with `name:`
plus a system-prompt body). Keep each persona repo-specific — customized from
references (e.g. agency-agents), not installed wholesale.

When personas are `on` and the Claude adapter is enabled, `gyroscope init`/`check
--fix` **register** each valid persona by mirroring it byte-for-byte into
`.claude/agents/<name>.md` (that is where Claude scans for subagents; this
directory is the canonical, hub-routed source). The mirror is generated and
gyroscope-owned — don't hand-edit `.claude/agents/`; edit the persona here and
re-run. `check` flags a missing or drifted mirror.

<!-- gyroscope creates this directory and README as a blessed home for personas.
It does not ship personas itself; it copies persona bytes into .claude/agents/ for
registration but never authors persona content. -->


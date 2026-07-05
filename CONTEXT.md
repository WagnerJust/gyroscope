# CONTEXT.md

Shared vocabulary for **gyroscope**. A **glossary and nothing else** — canonical
terms and their precise meaning. No implementation details, no specs, no
decisions (those live in `docs/adr/`).

Agents: use these terms exactly. If you catch a term used two different ways,
stop and flag it.

## Terms

### The standard
gyroscope's fixed, opinionated set of agent docs written into a repo — the hub,
its spokes, the pointer files, and the enforcement hook, all at once. A *paved
road*, not a blank canvas.
_Avoid:_ "the config" (that only toggles spokes); "buckle" (the blank-canvas tool gyroscope evolved from).

### Hub
`AGENTS.md`: the single entry point every agent reads first. It routes to spokes
and holds no topic content of its own.
_Avoid:_ "index" / "README"; calling it a spoke.

### Spoke
One of the topic docs the hub routes to — `CONTEXT.md`, `docs/agents.md`,
`docs/adr/`, `docs/agents/`, `.local/local.md`. Each is optional and toggled in
`gyroscope.json`; the hub itself is always written.
_Avoid:_ pointer file (a spoke carries content; a pointer only redirects).

### Pointer file
A tool-specific file (`CLAUDE.md`, `GEMINI.md`) reduced to one routing line back
to the hub, so per-tool instructions cannot drift apart.
_Avoid:_ spoke (a pointer holds no content); hub.

### Routing line
The single canonical sentence written into every pointer file: "Before doing
anything else, read AGENTS.md and follow its routes."
_Avoid:_ "routing table" — the hub presents routes as a bulleted `## Routes` list.

### Doc-target registry
`internal/target`: the registry of the *many* tools that each get a pointer file.
Kept separate from enforcement because most tools cannot run hooks.
_Avoid:_ enforcement adapter (the other, separate registry).

### Enforcement adapter
`internal/enforce`: an adapter for one of the *few* harnesses that can run a hook
to make agents actually read the hub (Claude now, PI later).
_Avoid:_ doc target (far more tools, pointer-only, no enforcement).

### SessionStart hook
The Claude adapter's mechanism: a `.claude/settings.json` entry that `cat`s the
hub + `docs/agents.md` + `.local/local.md` into context at the start of each
session, so the rules are present without the agent choosing to open them.
_Avoid:_ pointer file (a hook enforces; a pointer only redirects).

### Scaffold
A template file the binary writes verbatim: structure plus double-curly-brace
placeholders and guidance comments, with no repo-specific content yet.
_Avoid:_ implying a templating engine — the binary copies embedded bytes, it does not render.

### Placeholder
A double-curly-brace marker inside a scaffold that the skill later replaces with
real, repo-specific content. None should survive in a filled, committed doc.

### The binary
`gyroscope`: the deterministic, non-interactive CLI. It guarantees *structure +
hook* and never fills placeholders.
_Avoid:_ conflating it with the skill (the skill supplies content).

### The skill
`/gyroscope`: the companion (embedded `skill/SKILL.md`, installed via
`install-skill`) that runs the interview and fills the scaffolds' placeholders.
User-invoked, not model-invoked.
_Avoid:_ conflating it with the binary (the binary writes structure).

### dry-run
`init`'s default mode: print the plan of what *would* be written and touch
nothing.

### apply
`init --apply`: actually write the standard, pointer files, and hook.

### force
`init --force`: overwrite files that already exist. Without it, `init --apply`
pre-flights every destination and refuses (all-or-nothing) if any collides.

### gyroscope.json
The optional repo-root config file that toggles which spokes are on. Absent means
every spoke on — the opinionated default.

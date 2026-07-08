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

### Managed region
The slice of the hub gyroscope owns and re-writes — the content between the
`<!-- gyroscope:managed -->` and `<!-- /gyroscope -->` markers (the routes, the
pointer-files list, the personas directive). Everything outside the markers is the
user's: gyroscope never writes it and `check` never reads it. Injecting or updating
the region in place (not clobbering the whole file) is a *merge* (see ADR 0007).
_Avoid:_ "managed block" and "managed region" used for different things — they are
the same region; pick "managed region." Not the whole hub (only the delimited slice).

### Spoke
One of the topic docs the hub routes to — `CONTEXT.md`, `docs/agents.md`,
`CONTRIBUTING.md`, `docs/adr/`, `docs/agents/`, `TODO.md`, `.local/local.md`. Each
is optional and toggled in `gyroscope.json` (which also prunes its hub route when
off); the hub itself is always written.
_Avoid:_ pointer file (a spoke carries content; a pointer only redirects); process artifact (written but not routed).

### Process artifact
A file gyroscope writes that Git or GitHub enforces directly rather than the hub
routing to it — `.github/pull_request_template.md` (`prTemplate`) and `.gitmessage`
(`commitConvention`). Part of the standard, but carries no hub route.
_Avoid:_ spoke (a spoke is hub-routed; a process artifact is applied by tooling).

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
hub + `docs/agents.md` + the state files (`TODO.md`, `.local/todo.md`) +
`.local/local.md` into context at the start of each session, so the rules and
current progress are present without the agent choosing to open them.
_Avoid:_ pointer file (a hook enforces; a pointer only redirects).

### Scaffold
A template file the binary writes verbatim: structure, sometimes with
double-curly-brace placeholders and guidance comments for the skill to fill. Some
scaffolds (`CONTRIBUTING.md`, `TODO.md`, `.local/todo.md`, the process artifacts)
ship ready-to-use with no placeholders.
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

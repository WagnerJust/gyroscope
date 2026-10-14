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
The slice of a file gyroscope owns and re-writes — the content between the
`<!-- gyroscope:managed -->` and `<!-- /gyroscope -->` markers. The hub's region
holds the routes, the pointer-files list, and the personas directive (config-
rendered); a spoke like `CONTRIBUTING.md` may also carry a region for a static
gyroscope-owned block (the contributor note, see ADR 0012). Everything outside the
markers is the user's: gyroscope never writes it and `check` never reads it.
Injecting or updating the region in place (not clobbering the whole file) is a
*merge* (see ADR 0007).
_Avoid:_ "managed block" and "managed region" used for different things — they are
the same region; pick "managed region." Not the whole file (only the delimited slice).

### Spoke
One of the topic docs the hub routes to — `CONTEXT.md`, `docs/agents.md`,
`CONTRIBUTING.md`, `docs/adr/`, `docs/agents/`, `TODO.md`, `DONE.md`,
`.local/local.md`. Each is optional and toggled in `gyroscope.json` (which also
prunes its hub route when off); the hub itself is always written. `TODO.md`
(open work) and `DONE.md` (completed archive) both belong to the `state` toggle,
but `DONE.md` is routed only — never injected by the hook (see ADR 0009).
_Avoid:_ pointer file (a spoke carries content; a pointer only redirects); process artifact (written but not routed).

### Process artifact
A file gyroscope writes that Git or GitHub enforces directly rather than the hub
routing to it — `.github/pull_request_template.md` (`prTemplate`) and `.gitmessage`
(`commitConvention`). Part of the standard, but carries no hub route.
_Avoid:_ spoke (a spoke is hub-routed; a process artifact is applied by tooling).

### Pointer file
A tool-specific file (`CLAUDE.md`, `GEMINI.md`) reduced to a redirect back to the
hub, so per-tool instructions cannot drift apart. Six carry the routing line;
`CLAUDE.md` carries an import (`@AGENTS.md`, plus `@docs/agents.md` when the
agents spoke is on), because Claude Code loads imported files in full at launch
but only *decides* whether to open a file it is told about in words (see ADR 0022).
_Avoid:_ spoke (a pointer holds no content); hub; "the routing line" for
`CLAUDE.md` (it is an import pointer).

### Routing line
The single canonical sentence written into every pointer file except `CLAUDE.md`:
"Before doing anything else, read AGENTS.md and follow its routes."
_Avoid:_ "routing table" — the hub presents routes as a bulleted `## Routes` list;
calling `CLAUDE.md`'s `@AGENTS.md` import a routing line (see ADR 0022).

### Doc-target registry
`internal/target`: the registry of the *many* tools that each get a pointer file.
Kept separate from enforcement because most tools cannot run hooks.
_Avoid:_ enforcement adapter (the other, separate registry).

### Enforcement adapter
`internal/enforce`: an adapter for one of the *few* harnesses that have a native
mechanism gyroscope can install — a hook, an extension, an always-apply rule, an
import — to make agents actually encounter the hub and the state (Claude, PI,
OpenCode, Codex, Cursor today).
_Avoid:_ doc target (far more tools, pointer-only, no enforcement).

### SessionStart hook
The Claude adapter's mechanism for *state*: a `.claude/settings.json` entry that
runs a budgeted POSIX loop at the start of each session, injecting the state files
(`gyroscope.json`, `TODO.md`, `.local/todo.md`, `.local/local.md`, in that priority
order) while they fit the state budget, and printing a one-line pointer for each
that does not. Rules (`AGENTS.md`, `docs/agents.md`) do not travel this way —
Claude loads them through `CLAUDE.md`'s imports (see ADR 0022). `DONE.md` is a
spoke and is routed by the hub, but is deliberately never injected (see ADR 0009).
The command carries a `# gyroscope-hook` tag so a re-install replaces it rather
than appending a second.
_Avoid:_ pointer file (a hook enforces; a pointer only redirects); "pure cat" (it
was, until ADR 0022); describing it as injecting the hub.

### State budget
The 9,000-byte ceiling the SessionStart hook spends on state files, in priority
order, so its whole output stays under Claude Code's 10,000-character cap on hook
output — past which only a 2,000-character preview arrives. `check` simulates the
same budget: a `TODO.md` that cannot fit on its own is drift; a file deferred
behind others is a note (see ADR 0022).
_Avoid:_ "the cap" (that is the harness's limit; the budget is gyroscope's margin
under it); confusing bytes (`wc -c`, what the hook counts) with characters (what
the harness counts) — bytes are the conservative side.

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
hook + persona registration* and never fills placeholders. It **copies** persona
bytes (the persona mirror) but never **authors** persona content — copying bytes is
not rendering; the skill authors, the binary mirrors (see ADR 0010).
_Avoid:_ conflating it with the skill (the skill supplies content); "the binary
never touches personas" (it registers them by byte copy, it just never writes their
content).

### Persona mirror
The byte-for-byte copy of each valid `docs/agents/` persona into
`.claude/agents/<name>.md` that makes Claude register it as a dispatchable subagent.
`docs/agents/` stays the canonical, hub-routed source; the mirror is a generated,
gyroscope-owned file (overwritten on drift, atomically) — the one deliberate
exception to `WriteGuarded`'s refuse-to-clobber. Runs only when
`spokes.personas == on` AND `enforce.claude` (Claude-specific). A *valid* persona is
a `docs/agents/*.md` other than `README.md` whose frontmatter carries a `name:`; the
dest filename is that `name:` (see ADR 0010).
_Avoid:_ "symlink" (it is a copy — drift caught by check); implying the binary
authors the content (it copies bytes the skill authored).

### The skill
`/gyroscope`: the companion (embedded `skill/SKILL.md`, installed via
`install-skill`) that runs the interview, reconciles an existing repo, and fills the
scaffolds' placeholders. Model-invocable — a plain request like "get this repo up to
date with gyroscope" can trigger it — but gated: it never writes without showing a
plan and getting approval (see ADR 0008).
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

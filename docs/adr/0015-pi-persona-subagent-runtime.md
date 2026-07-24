# ADR 0015: PI persona parity is a generated runtime extension, not a file mirror

- **Status:** accepted
- **Date:** 2026-07-23

## Context

ADR 0014 registers a repo's personas as subagents for every harness that has a
native, file-based subagent registry (Claude, OpenCode, Codex) by re-encoding each
persona into that registry's format. PI is the one enabled harness with **no native
subagent registry** — by design. Its README states: *"No sub-agents. There's many
ways to do this… build your own with extensions."* There is no `.pi/agents/`
directory to write, so the mirror mechanism cannot give PI persona parity.

But PI's extension API *can* register an agent-callable tool at runtime
(`pi.registerTool({ name, label, description, parameters, execute })`), and PI's own
reference `subagent` extension shows the intended pattern: the tool's `execute`
spawns a fresh **headless `pi`** process (`child_process.spawn`) with a restricted
system prompt. So parity for PI is *buildable* — as runtime behavior, not a copied
file. gyroscope already installs a PI extension for hub-spoke injection (ADR 0006),
so a second generated extension is a natural extension of that mechanism.

## Decision

When `personas == on` and `enforce.pi` is set, gyroscope generates
`.pi/extensions/gyroscope-personas.ts`, which registers one subagent tool per
persona. Each tool's `execute` spawns a headless `pi -p --no-session
--append-system-prompt <persona body> "Task: <task>"` and returns its printed
response. Key choices:

- **gyroscope bakes the metadata; the extension never parses YAML.** The generator
  parses each persona once (the same parser the file mirrors use) and bakes a JSON
  array of `{name, label, description, prompt}` into the extension. This avoids a
  second, drifting YAML parser in TypeScript — the single-parser rule from ADR
  0010/0014.
- **It rides the ordinary mirror pipeline.** The extension is one gyroscope-owned
  generated file, so it is emitted as a `persona.Mirror` — written by the same
  overwrite-on-drift writer and verified byte-for-byte by `check`, exactly like the
  file mirrors. It is *one* file for all personas (not one per persona), because a PI
  extension is an aggregate registrar, not a per-agent record.
- **Plain `-p`, not `--mode json`.** Headless `pi -p` prints the response as text, so
  `execute` returns stdout directly — avoiding a brittle, guessed JSON-event parser.
- **Tool restriction is a documented gap, not a silent one.** A persona's `tools:`
  field (Claude tool names) is deliberately NOT mapped to PI's `--tools`: the names
  do not map cleanly, and a wrong name would fail the whole run. The subagent runs
  with PI's default toolset; the extension header says so.

## Consequences

- PI reaches persona-*dispatch* parity: each authored persona becomes a tool the
  primary PI agent can call, running that persona's instructions in isolation — from
  the same canonical `docs/agents/` source as every other harness.
- It is heavier than a mirror: gyroscope now generates executable TypeScript, and its
  correctness depends on PI's extension API (`registerTool`) and headless flags
  (`-p`, `--no-session`, `--append-system-prompt`) staying stable. Those are pinned to
  what PI's repo documents and its reference `subagent` example uses as of this ADR.
- Tool-restriction fidelity is the known gap; if a Claude→PI tool-name map is ever
  wanted, it slots into the generator without changing this shape.
- Parity is still gated and opt-in (`enforce.pi` + `personas == on`), consistent with
  ADR 0002/0006 — a repo not driven by PI generates nothing.

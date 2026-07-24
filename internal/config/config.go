// Package config reads gyroscope.json — a repo-level file toggling which spokes
// of the standard are enabled. Absent file means "all spokes on" (the opinionated
// default). Parsing over a Default() base means an omitted spoke stays on; only a
// spoke explicitly set false is turned off.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	Spokes  SpokeSet   `json:"spokes"`
	Enforce EnforceSet `json:"enforce"`
	// Custom lists extra spoke doc files beyond the built-in five. Absent means
	// none; the writer skips any entry missing a Name or Dest.
	Custom []CustomSpoke `json:"custom"`
	// Sync is opt-in (default false): when set, gyroscope generates a committed,
	// dependency-free `gyroscope-sync.sh` that reconstructs the persona-registration
	// mirrors for a teammate who cloned a repo whose generated dirs (.claude/, etc.)
	// are gitignored and never installed gyroscope. Idempotent (skip-if-exists), so
	// it "syncs if it hasn't been." The alternative to committing the generated dirs
	// (the ship-to-team note). See ADR 0018.
	Sync bool `json:"sync"`
	// CodeReview is an opt-in persona PRESET (default false): instead of the
	// from-scratch persona interview, the /gyroscope skill scaffolds a curated
	// multi-persona code-review bundle — layer-aligned reviewer personas
	// (repo-tuned), a synthesizer, an intent-verifier grounded against the repo's
	// decision record, a code-review orchestrator runbook, and the grounding ADR.
	// The binary only records the flag; the skill authors the content (binary/skill
	// split). See ADR 0017.
	CodeReview bool `json:"codeReview"`
}

// EnforceSet toggles which harness enforcement adapters init installs and check
// verifies. Claude is default-on (unchanged behavior); PI is opt-in because its
// extension is more intrusive and only useful in repos actually driven by PI
// (see ADR 0002 — enforcement is opt-in per harness).
type EnforceSet struct {
	Claude bool `json:"claude"`
	PI     bool `json:"pi"`
	// OpenCode and Codex are opt-in (default false), like PI: both read AGENTS.md
	// natively (no hook to install), so enabling them wires the parts that are NOT
	// free — mirroring docs/agents/ personas into each tool's native subagent
	// registry (.opencode/agents/, .codex/agents/) and, for Codex, the
	// commit_attribution setting. Off means gyroscope writes nothing tool-specific.
	OpenCode bool `json:"opencode"`
	Codex    bool `json:"codex"`
	// Cursor is opt-in (default false): it installs a .cursor/rules/ always-apply
	// rule that force-injects the hub directive into every Cursor request — the
	// enforcement analog to Claude's hook, above the passive .cursorrules pointer.
	// (Zed needs no adapter: it natively reads .rules every session.)
	Cursor bool `json:"cursor"`
	// AIAttribution controls whether AI agents may sign their work — the
	// `Co-Authored-By` trailer / "Generated with …" footer on commits and PRs/MRs.
	// Default true (unchanged behavior). Set false to suppress it repo-wide: the
	// Claude adapter writes `includeCoAuthoredBy:false` into .claude/settings.json,
	// and the hub carries a standing directive so other harnesses honor it too.
	AIAttribution bool `json:"aiAttribution"`
}

// CustomSpoke is a user-defined spoke: a doc file gyroscope scaffolds and routes
// from the hub. The route is auto-generated from Name and Dest.
type CustomSpoke struct {
	Name string `json:"name"`
	Dest string `json:"dest"`
}

// PersonaState is the lifecycle of the docs/agents/ personas spoke. Only
// PersonaUnknown triggers the hub's "ask the user" directive; the rest are silent.
// It unmarshals from either the legacy bool (true→unknown, false→off) or a string.
type PersonaState string

const (
	PersonaUnknown PersonaState = "unknown"
	PersonaOn      PersonaState = "on"
	PersonaSkipped PersonaState = "skipped"
	PersonaOff     PersonaState = "off"
)

// Enabled reports whether the personas spoke exists at all (route + README). Only
// PersonaOff drops it.
func (p PersonaState) Enabled() bool { return p != PersonaOff }

// IsUnknown reports whether the persona decision is still pending (nudge state).
func (p PersonaState) IsUnknown() bool { return p == PersonaUnknown }

// Valid reports whether p is one of the four known states.
func (p PersonaState) Valid() bool {
	switch p {
	case PersonaUnknown, PersonaOn, PersonaSkipped, PersonaOff:
		return true
	}
	return false
}

func (p *PersonaState) UnmarshalJSON(b []byte) error {
	var boolVal bool
	if err := json.Unmarshal(b, &boolVal); err == nil {
		if boolVal {
			*p = PersonaUnknown
		} else {
			*p = PersonaOff
		}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("personas: expected bool or string, got %s", b)
	}
	if !PersonaState(s).Valid() {
		return fmt.Errorf("personas: unknown state %q (want unknown|on|skipped|off)", s)
	}
	*p = PersonaState(s)
	return nil
}

type SpokeSet struct {
	Context  bool         `json:"context"`
	Agents   bool         `json:"agents"`
	ADR      bool         `json:"adr"`
	Personas PersonaState `json:"personas"`
	Local    bool         `json:"local"`
	// Contributing is a hub-routed spoke covering the contribution process; it
	// defers to the agents spoke for conventions rather than duplicating them.
	Contributing bool `json:"contributing"`
	// State writes the progress files a fresh session resumes from: a tracked,
	// repo-wide TODO.md and a gitignored, personal .local/todo.md. Both are
	// injected by the SessionStart hook so a new chat picks up where the last
	// one left off.
	State bool `json:"state"`
	// PRTemplate and CommitConvention encode commit/PR judgement but are applied
	// by tooling (Git's commit template, GitHub's PR form), not read via the hub
	// — they carry no route.
	PRTemplate       bool `json:"prTemplate"`
	CommitConvention bool `json:"commitConvention"`
}

// Default is the opinionated standard: every spoke on.
func Default() Config {
	return Config{
		Spokes: SpokeSet{
			Context: true, Agents: true, ADR: true, Personas: PersonaUnknown, Local: true,
			Contributing: true, PRTemplate: true, CommitConvention: true, State: true,
		},
		Enforce: EnforceSet{Claude: true, PI: false, OpenCode: false, Codex: false, Cursor: false, AIAttribution: true},
	}
}

// Load reads gyroscope.json from dir; a missing file yields Default().
func Load(dir string) (Config, error) {
	b, err := os.ReadFile(filepath.Join(dir, "gyroscope.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", filepath.Join(dir, "gyroscope.json"), err)
	}
	return cfg, nil
}

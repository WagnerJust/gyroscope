package main

import (
	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/persona"
)

// personaMirrorGated reports whether gyroscope should mirror docs/agents/
// personas into .claude/agents/ for cfg. The mirror is Claude-specific (it makes
// Claude register the personas as subagents), so it runs only when the personas
// spoke is actually wired (`personas == on`) AND the Claude enforcement adapter is
// enabled. It is deliberately narrower than PersonaState.Enabled(): an unknown,
// skipped, or off persona spoke — and a repo whose Claude adapter is off (e.g. a
// PI-only repo, which has its own agent mechanism) — mirrors nothing.
func personaMirrorGated(cfg config.Config) bool {
	return cfg.Spokes.Personas == config.PersonaOn && cfg.Enforce.Claude
}

// planPersonaMirrors returns the persona mirrors init would write for repoDir when
// gated, else an empty slice. Sharing the gate + Plan keeps init's dry-run listing,
// init --apply, and check's verification in agreement about which personas register.
func planPersonaMirrors(repoDir string, cfg config.Config) ([]persona.Mirror, error) {
	if !personaMirrorGated(cfg) {
		return nil, nil
	}
	return persona.Plan(repoDir)
}

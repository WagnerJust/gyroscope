package main

import (
	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/persona"
)

// personaTargets returns the harness subagent registries gyroscope should mirror
// docs/agents/ personas into for cfg. A registry is included only when the personas
// spoke is actually wired (`personas == on`) AND that harness's enforcement adapter
// is enabled — deliberately narrower than PersonaState.Enabled(): an unknown,
// skipped, or off persona spoke mirrors nothing. Claude, OpenCode, and Codex each
// have a native, file-based subagent registry the binary can write; PI has none by
// design (its parity is an extension-built runtime, not a mirror — ADR 0006), so it
// is absent here.
func personaTargets(cfg config.Config) []persona.Target {
	if cfg.Spokes.Personas != config.PersonaOn {
		return nil
	}
	var ts []persona.Target
	if cfg.Enforce.Claude {
		ts = append(ts, persona.Claude)
	}
	if cfg.Enforce.OpenCode {
		ts = append(ts, persona.OpenCode)
	}
	if cfg.Enforce.Codex {
		ts = append(ts, persona.Codex)
	}
	return ts
}

// planPersonaMirrors returns the persona mirrors init would write for repoDir —
// the union across every enabled harness registry. Sharing personaTargets + PlanFor
// keeps init's dry-run listing, init --apply, and check's verification in agreement
// about which personas register where.
func planPersonaMirrors(repoDir string, cfg config.Config) ([]persona.Mirror, error) {
	var all []persona.Mirror
	for _, t := range personaTargets(cfg) {
		ms, err := persona.PlanFor(repoDir, t)
		if err != nil {
			return nil, err
		}
		all = append(all, ms...)
	}
	return all, nil
}

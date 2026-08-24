package main

import (
	"github.com/WagnerJust/gyroscope/internal/command"
	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/persona"
)

// commandTargets returns the harness command registries gyroscope re-encodes
// docs/commands/ into for cfg — each enabled harness. Unlike personas, commands are
// NOT gated on a persona state: any repo with docs/commands/*.md gets them installed
// for its enabled harnesses (a missing docs/commands/ makes each target a no-op).
func commandTargets(cfg config.Config) []command.Target {
	var ts []command.Target
	if cfg.Enforce.Claude {
		ts = append(ts, command.Claude)
	}
	if cfg.Enforce.OpenCode {
		ts = append(ts, command.OpenCode)
	}
	if cfg.Enforce.PI {
		ts = append(ts, command.PI)
	}
	return ts
}

// planCommandMirrors returns the command mirrors init would write for repoDir — the
// union across every enabled harness registry.
func planCommandMirrors(repoDir string, cfg config.Config) ([]persona.Mirror, error) {
	var all []persona.Mirror
	for _, t := range commandTargets(cfg) {
		ms, err := command.PlanFor(repoDir, t)
		if err != nil {
			return nil, err
		}
		all = append(all, ms...)
	}
	return all, nil
}

// planMirrors returns every gyroscope-owned generated file for cfg — persona
// registrations plus command re-encodings — as one list. init writes it, check
// verifies each byte-for-byte, and the sync script bakes it, so all three artifact
// flows agree on the full set of generated files.
func planMirrors(repoDir string, cfg config.Config) ([]persona.Mirror, error) {
	p, err := planPersonaMirrors(repoDir, cfg)
	if err != nil {
		return nil, err
	}
	c, err := planCommandMirrors(repoDir, cfg)
	if err != nil {
		return nil, err
	}
	m, err := planMCPMirrors(repoDir, cfg)
	if err != nil {
		return nil, err
	}
	return append(append(p, c...), m...), nil
}

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
	Spokes SpokeSet `json:"spokes"`
	// Custom lists extra spoke doc files beyond the built-in five. Absent means
	// none; the writer skips any entry missing a Name or Dest.
	Custom []CustomSpoke `json:"custom"`
}

// CustomSpoke is a user-defined spoke: a doc file gyroscope scaffolds and routes
// from the hub. The route is auto-generated from Name and Dest.
type CustomSpoke struct {
	Name string `json:"name"`
	Dest string `json:"dest"`
}

type SpokeSet struct {
	Context  bool `json:"context"`
	Agents   bool `json:"agents"`
	ADR      bool `json:"adr"`
	Personas bool `json:"personas"`
	Local    bool `json:"local"`
	// Contributing is a hub-routed spoke covering the contribution process; it
	// defers to the agents spoke for conventions rather than duplicating them.
	Contributing bool `json:"contributing"`
	// PRTemplate and CommitConvention encode commit/PR judgement but are applied
	// by tooling (Git's commit template, GitHub's PR form), not read via the hub
	// — they carry no route.
	PRTemplate       bool `json:"prTemplate"`
	CommitConvention bool `json:"commitConvention"`
}

// Default is the opinionated standard: every spoke on.
func Default() Config {
	return Config{Spokes: SpokeSet{
		Context: true, Agents: true, ADR: true, Personas: true, Local: true,
		Contributing: true, PRTemplate: true, CommitConvention: true,
	}}
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

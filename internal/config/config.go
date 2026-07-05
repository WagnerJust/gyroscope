// Package config reads gyroscope.json — a repo-level file toggling which spokes
// of the standard are enabled. Absent file means "all spokes on" (the opinionated
// default). Parsing over a Default() base means an omitted spoke stays on; only a
// spoke explicitly set false is turned off.
package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	Spokes SpokeSet `json:"spokes"`
}

type SpokeSet struct {
	Context  bool `json:"context"`
	Agents   bool `json:"agents"`
	ADR      bool `json:"adr"`
	Personas bool `json:"personas"`
	Local    bool `json:"local"`
}

// Default is the opinionated standard: every spoke on.
func Default() Config {
	return Config{Spokes: SpokeSet{Context: true, Agents: true, ADR: true, Personas: true, Local: true}}
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
		return Config{}, err
	}
	return cfg, nil
}

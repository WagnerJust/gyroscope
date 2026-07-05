package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAbsentReturnsDefaultAllOn(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Spokes != Default().Spokes || cfg.Custom != nil {
		t.Fatalf("absent config should be Default(), got %+v", cfg)
	}
	if !cfg.Spokes.Local {
		t.Fatal("local spoke should default on")
	}
}

func TestLoadOverridesOnlyNamedSpoke(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"spokes":{"local":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Spokes.Local {
		t.Fatal("local should be off")
	}
	if !cfg.Spokes.Context {
		t.Fatal("unnamed spokes must stay on")
	}
}

func TestLoadCustomSpokes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"custom":[{"name":"Security","dest":"docs/security.md"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Custom) != 1 {
		t.Fatalf("expected 1 custom spoke, got %d", len(cfg.Custom))
	}
	if cfg.Custom[0].Name != "Security" || cfg.Custom[0].Dest != "docs/security.md" {
		t.Fatalf("unexpected custom spoke: %+v", cfg.Custom[0])
	}
	// Named spoke config absent → spokes still default on.
	if !cfg.Spokes.Context {
		t.Fatal("unnamed spokes must stay on")
	}
}

func TestLoadAbsentCustomIsNil(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"spokes":{"local":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Custom != nil {
		t.Fatalf("absent custom should be nil, got %+v", cfg.Custom)
	}
}

func TestLoadMalformedJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"spokes":`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err == nil {
		t.Fatal("malformed JSON should return a non-nil error")
	}
	if !strings.Contains(err.Error(), "gyroscope.json") {
		t.Fatalf("malformed JSON error should name the file, got %q", err.Error())
	}
	if cfg.Spokes != (SpokeSet{}) || cfg.Custom != nil {
		t.Fatalf("malformed JSON should return zero Config{}, got %+v", cfg)
	}
}

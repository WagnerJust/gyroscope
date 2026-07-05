package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAbsentReturnsDefaultAllOn(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg != Default() {
		t.Fatalf("absent config should be Default(), got %+v", cfg)
	}
	if !cfg.Spokes.Local {
		t.Fatal("local spoke should default on")
	}
}

func TestLoadOverridesOnlyNamedSpoke(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"spokes":{"local":false}}`), 0o644)
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

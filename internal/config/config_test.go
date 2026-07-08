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

func TestDefaultEnforceClaudeOnPIOff(t *testing.T) {
	e := Default().Enforce
	if !e.Claude || e.PI {
		t.Fatalf("default enforce should be claude-on/pi-off, got %+v", e)
	}
}

func TestEnforcePIOptInKeepsClaude(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"enforce":{"pi":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enforce.PI {
		t.Fatal("pi should be enabled")
	}
	if !cfg.Enforce.Claude {
		t.Fatal("claude should stay on when only pi is named")
	}
}

func TestEnforceAbsentKeepsDefaults(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enforce != (EnforceSet{Claude: true, PI: false}) {
		t.Fatalf("absent enforce should be default, got %+v", cfg.Enforce)
	}
}

func TestPersonasBoolBackCompat(t *testing.T) {
	cases := map[string]PersonaState{
		`{"spokes":{"personas":true}}`:  PersonaUnknown,
		`{"spokes":{"personas":false}}`: PersonaOff,
	}
	for body, want := range cases {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := Load(dir)
		if err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		if cfg.Spokes.Personas != want {
			t.Fatalf("%s: got %q, want %q", body, cfg.Spokes.Personas, want)
		}
	}
}

func TestPersonasStringStates(t *testing.T) {
	for _, want := range []PersonaState{PersonaUnknown, PersonaOn, PersonaSkipped, PersonaOff} {
		dir := t.TempDir()
		body := `{"spokes":{"personas":"` + string(want) + `"}}`
		if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := Load(dir)
		if err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		if cfg.Spokes.Personas != want {
			t.Fatalf("got %q, want %q", cfg.Spokes.Personas, want)
		}
	}
}

func TestPersonasBadStateWrapsFilename(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gyroscope.json"), []byte(`{"spokes":{"personas":"wat"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for unknown persona state")
	}
	if !strings.Contains(err.Error(), "gyroscope.json") || !strings.Contains(err.Error(), "wat") {
		t.Fatalf("error should name the file and the bad value: %v", err)
	}
}

func TestDefaultPersonasUnknown(t *testing.T) {
	if Default().Spokes.Personas != PersonaUnknown {
		t.Fatalf("default personas should be unknown, got %q", Default().Spokes.Personas)
	}
}

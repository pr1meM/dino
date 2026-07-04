package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileYieldsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg != Default() {
		t.Fatalf("cfg = %+v, want defaults", cfg)
	}
}

func TestLoadOverridesTabSize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"tabSize": 2, "useSpaces": false}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TabSize != 2 || cfg.UseSpaces != false {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestLoadRejectsInvalidTabSize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"tabSize": 0}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TabSize != Default().TabSize {
		t.Fatalf("tabSize = %d, want default %d", cfg.TabSize, Default().TabSize)
	}
}

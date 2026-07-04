// Package config loads user-tunable settings for dino from a small
// JSON file, e.g. ~/.config/dino/config.json:
//
//	{ "tabSize": 4, "useSpaces": true }
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	TabSize   int  `json:"tabSize"`
	UseSpaces bool `json:"useSpaces"`
}

func Default() Config {
	return Config{TabSize: 4, UseSpaces: true}
}

// Path returns the default config file location, honoring XDG_CONFIG_HOME.
func Path() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "dino", "config.json")
}

// Load reads the config file at path, overlaying it onto the defaults.
// A missing file is not an error; it just yields Default().
func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.TabSize <= 0 {
		cfg.TabSize = Default().TabSize
	}
	return cfg, nil
}

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SourceConfig points the app at one browsable root directory.
// Local folders and mounted network drives (e.g. autobutler) are both
// just filesystem roots, so they share this same config shape.
type SourceConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Config struct {
	Sources []SourceConfig `json:"sources"`
}

// DefaultPath returns the config file location, creating no files itself.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "goodmorning-photos", "config.json"), nil
}

// Load reads the config at path. If the file doesn't exist, it returns a
// best-effort default (the user's Pictures folder, or cwd as a last resort)
// instead of failing, so the app is usable with zero setup.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", path, err)
	}
	if len(cfg.Sources) == 0 {
		return Config{}, fmt.Errorf("config %s has no sources configured", path)
	}
	return cfg, nil
}

// Save writes cfg to path as indented JSON, creating parent directories.
func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func defaultConfig() Config {
	if home, err := os.UserHomeDir(); err == nil {
		pictures := filepath.Join(home, "Pictures")
		if info, err := os.Stat(pictures); err == nil && info.IsDir() {
			return Config{Sources: []SourceConfig{{Name: "Pictures", Path: pictures}}}
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return Config{Sources: []SourceConfig{{Name: "Current Directory", Path: cwd}}}
}

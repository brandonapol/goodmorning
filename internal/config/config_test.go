package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingFileFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	pictures := filepath.Join(dir, "Pictures")
	if err := os.MkdirAll(pictures, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(filepath.Join(dir, "does-not-exist.json"))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if len(cfg.Sources) != 1 {
		t.Fatalf("Sources = %v, want exactly one default source", cfg.Sources)
	}
	if cfg.Sources[0].Path != pictures {
		t.Errorf("Sources[0].Path = %q, want %q", cfg.Sources[0].Path, pictures)
	}
}

func TestLoad_MissingFileFallsBackToCwdWhenNoPictures(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir) // no Pictures subdir created

	cfg, err := Load(filepath.Join(dir, "does-not-exist.json"))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if len(cfg.Sources) != 1 {
		t.Fatalf("Sources = %v, want exactly one default source", cfg.Sources)
	}
	if cfg.Sources[0].Name != "Current Directory" {
		t.Errorf("Sources[0].Name = %q, want %q", cfg.Sources[0].Name, "Current Directory")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	want := Config{Sources: []SourceConfig{
		{Name: "Pictures", Path: "/home/me/Pictures"},
		{Name: "Autobutler", Path: "/mnt/autobutler/photos"},
	}}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.Sources) != len(want.Sources) {
		t.Fatalf("Sources = %v, want %v", got.Sources, want.Sources)
	}
	for i := range want.Sources {
		if got.Sources[i] != want.Sources[i] {
			t.Errorf("Sources[%d] = %v, want %v", i, got.Sources[i], want.Sources[i])
		}
	}
}

func TestLoad_EmptySourcesIsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := Save(path, Config{Sources: nil}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Error("Load() with zero sources: want error, got nil")
	}
}

func TestLoad_MalformedJSONIsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Error("Load() with malformed JSON: want error, got nil")
	}
}

func TestSave_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deeper", "config.json")
	cfg := Config{Sources: []SourceConfig{{Name: "X", Path: "/x"}}}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("DefaultPath() = %q, want basename config.json", path)
	}
	if filepath.Base(filepath.Dir(path)) != "goodmorning-photos" {
		t.Errorf("DefaultPath() = %q, want parent dir goodmorning-photos", path)
	}
}

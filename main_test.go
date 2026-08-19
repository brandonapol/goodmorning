package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brandonapol/goodmorning/internal/config"
)

func TestBuildSources_AdhocPathIgnoresConfig(t *testing.T) {
	dir := t.TempDir()
	sources, err := buildSources("/nonexistent/config.json", dir)
	if err != nil {
		t.Fatalf("buildSources() error = %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("sources = %v, want exactly 1", sources)
	}
	if sources[0].Name() != "Photos" {
		t.Errorf("sources[0].Name() = %q, want %q", sources[0].Name(), "Photos")
	}
}

func TestBuildSources_ExplicitConfigPath(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	if err := config.Save(cfgPath, config.Config{Sources: []config.SourceConfig{
		{Name: "Pics", Path: "/photos/pics"},
		{Name: "NAS", Path: "/mnt/nas"},
	}}); err != nil {
		t.Fatal(err)
	}

	sources, err := buildSources(cfgPath, "")
	if err != nil {
		t.Fatalf("buildSources() error = %v", err)
	}
	if len(sources) != 2 {
		t.Fatalf("sources = %v, want 2", sources)
	}
	if sources[0].Name() != "Pics" || sources[1].Name() != "NAS" {
		t.Errorf("sources = [%q, %q], want [Pics, NAS]", sources[0].Name(), sources[1].Name())
	}
}

func TestBuildSources_MissingConfigFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir) // no Pictures subdir, so default is cwd

	sources, err := buildSources(filepath.Join(dir, "missing-config.json"), "")
	if err != nil {
		t.Fatalf("buildSources() error = %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("sources = %v, want exactly 1 fallback source", sources)
	}
}

func TestBuildSources_MalformedConfigIsError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := buildSources(cfgPath, ""); err == nil {
		t.Error("buildSources() with malformed config: want error, got nil")
	}
}

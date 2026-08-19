package render

import (
	"strings"
	"testing"
)

func newTestRenderer(t *testing.T) *Renderer {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	return NewRenderer()
}

func TestRender_ProducesNonEmptyArt(t *testing.T) {
	r := newTestRenderer(t)
	path := newTestImage(t)

	art, err := r.Render(path, 20)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.TrimSpace(art) == "" {
		t.Error("Render() returned empty art")
	}
}

func TestRender_ZeroWidthFallsBackToDefault(t *testing.T) {
	r := newTestRenderer(t)
	path := newTestImage(t)

	art, err := r.Render(path, 0)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.TrimSpace(art) == "" {
		t.Error("Render() with width=0 returned empty art")
	}
}

func TestRender_NegativeWidthFallsBackToDefault(t *testing.T) {
	r := newTestRenderer(t)
	path := newTestImage(t)

	if _, err := r.Render(path, -5); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
}

func TestRender_UsesCacheOnSecondCall(t *testing.T) {
	r := newTestRenderer(t)
	path := newTestImage(t)

	first, err := r.Render(path, 20)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	second, err := r.Render(path, 20)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if first != second {
		t.Errorf("Render() not stable across calls:\nfirst:  %q\nsecond: %q", first, second)
	}
}

func TestRender_MissingFileReturnsError(t *testing.T) {
	r := newTestRenderer(t)
	if _, err := r.Render("/does/not/exist.png", 20); err == nil {
		t.Error("Render() on missing file: want error, got nil")
	}
}

func TestRender_UnsupportedFileReturnsError(t *testing.T) {
	r := newTestRenderer(t)
	path := t.TempDir() + "/not-an-image.txt"
	if err := writeGarbageFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Render(path, 20); err == nil {
		t.Error("Render() on non-image file: want error, got nil")
	}
}

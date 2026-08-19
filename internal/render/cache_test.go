package render

import (
	"os"
	"testing"
	"time"
)

func newTestCache(t *testing.T) *diskCache {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	c, err := newDiskCache()
	if err != nil {
		t.Fatalf("newDiskCache() error = %v", err)
	}
	return c
}

func TestDiskCache_MissThenHit(t *testing.T) {
	c := newTestCache(t)
	path := newTestImage(t)

	if _, ok := c.get(path, 40, true); ok {
		t.Fatal("get() on empty cache: want miss, got hit")
	}

	c.put(path, 40, true, "the-art")

	got, ok := c.get(path, 40, true)
	if !ok {
		t.Fatal("get() after put: want hit, got miss")
	}
	if got != "the-art" {
		t.Errorf("get() = %q, want %q", got, "the-art")
	}
}

func TestDiskCache_KeyIncludesWidthAndColor(t *testing.T) {
	c := newTestCache(t)
	path := newTestImage(t)

	c.put(path, 40, true, "colored-40")

	if _, ok := c.get(path, 80, true); ok {
		t.Error("get() with different width: want miss, got hit")
	}
	if _, ok := c.get(path, 40, false); ok {
		t.Error("get() with different color mode: want miss, got hit")
	}

	got, ok := c.get(path, 40, true)
	if !ok || got != "colored-40" {
		t.Errorf("get() with matching key = (%q, %v), want (%q, true)", got, ok, "colored-40")
	}
}

func TestDiskCache_InvalidatedWhenFileChanges(t *testing.T) {
	c := newTestCache(t)
	path := newTestImage(t)

	c.put(path, 40, true, "stale-art")

	// Force a distinct mtime: some filesystems have coarse mtime
	// resolution, so bump it explicitly rather than just re-writing.
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}

	if _, ok := c.get(path, 40, true); ok {
		t.Error("get() after source file mtime changed: want miss, got hit (stale cache)")
	}
}

func TestDiskCache_MissingSourceFile(t *testing.T) {
	c := newTestCache(t)
	if _, ok := c.get("/does/not/exist.png", 40, true); ok {
		t.Error("get() for nonexistent source file: want miss, got hit")
	}
}

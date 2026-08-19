package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// writeFile creates path (and its parent dirs) with arbitrary content.
// Album/photo discovery only inspects file extensions, so the content
// itself doesn't need to be a valid image.
func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fake image bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newTestTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "loose.jpg"))
	writeFile(t, filepath.Join(root, "Vacation", "beach.jpg"))
	writeFile(t, filepath.Join(root, "Vacation", "mountain.PNG")) // extension case shouldn't matter
	writeFile(t, filepath.Join(root, "Vacation", "notes.txt"))    // non-image, must be ignored
	writeFile(t, filepath.Join(root, "Family", "dinner.jpg"))
	writeFile(t, filepath.Join(root, "Family", "iphone.heic")) // unsupported for now, must be ignored
	writeFile(t, filepath.Join(root, "Empty", "readme.txt"))   // dir with no images: not an album
	writeFile(t, filepath.Join(root, ".hidden", "secret.jpg")) // hidden dir: not an album
	return root
}

func TestListAlbums(t *testing.T) {
	src := NewFilesystemSource("Test", newTestTree(t))
	albums, err := src.ListAlbums(context.Background())
	if err != nil {
		t.Fatalf("ListAlbums() error = %v", err)
	}

	if len(albums) == 0 || albums[0].ID != AllPhotosAlbumID {
		t.Fatalf("albums[0] = %v, want the synthetic All Photos album first", albums)
	}

	var names []string
	for _, a := range albums[1:] {
		names = append(names, a.Name)
	}
	want := []string{"Family", "Vacation"}
	if len(names) != len(want) {
		t.Fatalf("album names = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("album[%d] = %q, want %q (Empty and .hidden must be excluded)", i, names[i], want[i])
		}
	}
}

func TestListPhotos_AllPhotosIsRecursiveAndSorted(t *testing.T) {
	src := NewFilesystemSource("Test", newTestTree(t))
	photos, err := src.ListPhotos(context.Background(), AllPhotosAlbumID)
	if err != nil {
		t.Fatalf("ListPhotos() error = %v", err)
	}

	var ids []string
	for _, p := range photos {
		ids = append(ids, p.ID)
	}

	for _, unwanted := range []string{"Vacation/notes.txt", "Family/iphone.heic", "Empty/readme.txt", ".hidden/secret.jpg"} {
		for _, id := range ids {
			if id == unwanted {
				t.Errorf("ListPhotos() included %q, want it excluded", unwanted)
			}
		}
	}

	for i := 1; i < len(ids); i++ {
		if ids[i-1] > ids[i] {
			t.Errorf("ListPhotos() not sorted: %q came before %q", ids[i-1], ids[i])
		}
	}

	if len(ids) != 4 {
		t.Errorf("got %d photos %v, want 4 (loose.jpg, Vacation/beach.jpg, Vacation/mountain.PNG, Family/dinner.jpg)", len(ids), ids)
	}
}

func TestListPhotos_ScopedToAlbum(t *testing.T) {
	src := NewFilesystemSource("Test", newTestTree(t))
	photos, err := src.ListPhotos(context.Background(), "Vacation")
	if err != nil {
		t.Fatalf("ListPhotos() error = %v", err)
	}
	if len(photos) != 2 {
		t.Fatalf("got %d photos, want 2 (beach.jpg, mountain.PNG): %v", len(photos), photos)
	}
	for _, p := range photos {
		if filepath.Dir(p.ID) != "Vacation" {
			t.Errorf("photo %v leaked from outside the Vacation album", p)
		}
	}
}

func TestLocalPath(t *testing.T) {
	root := newTestTree(t)
	src := NewFilesystemSource("Test", root)
	got, err := src.LocalPath(context.Background(), "Vacation/beach.jpg")
	if err != nil {
		t.Fatalf("LocalPath() error = %v", err)
	}
	want := filepath.Join(root, "Vacation", "beach.jpg")
	if got != want {
		t.Errorf("LocalPath() = %q, want %q", got, want)
	}
}

func TestListAlbums_ContextCanceled(t *testing.T) {
	src := NewFilesystemSource("Test", newTestTree(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := src.ListAlbums(ctx); err == nil {
		t.Error("ListAlbums() with canceled context: want error, got nil")
	}
}

func TestListPhotos_ContextCanceled(t *testing.T) {
	src := NewFilesystemSource("Test", newTestTree(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := src.ListPhotos(ctx, AllPhotosAlbumID); err == nil {
		t.Error("ListPhotos() with canceled context: want error, got nil")
	}
}

func TestListAlbums_MissingRoot(t *testing.T) {
	src := NewFilesystemSource("Test", filepath.Join(t.TempDir(), "does-not-exist"))
	if _, err := src.ListAlbums(context.Background()); err == nil {
		t.Error("ListAlbums() on missing root: want error, got nil")
	}
}

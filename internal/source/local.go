package source

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AllPhotosAlbumID is the synthetic album that lists every photo under a
// FilesystemSource's root, recursively.
const AllPhotosAlbumID = "__all__"

// supported image extensions for phase 1. HEIC is intentionally excluded
// for now (Go's stdlib can't decode it without extra tooling) and can be
// added later without changing this interface.
var imageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".bmp":  true,
	".tif":  true,
	".tiff": true,
}

// FilesystemSource treats a directory tree as a photo library: immediate
// subdirectories that contain images (recursively) become albums, plus a
// synthetic "All Photos" album. It backs both local folders and mounted
// network drives (e.g. autobutler) since both are just paths on disk.
type FilesystemSource struct {
	name string
	root string
}

func NewFilesystemSource(name, root string) *FilesystemSource {
	return &FilesystemSource{name: name, root: root}
}

func (f *FilesystemSource) Name() string { return f.name }

func (f *FilesystemSource) ListAlbums(ctx context.Context) ([]Album, error) {
	entries, err := os.ReadDir(f.root)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", f.root, err)
	}

	albums := []Album{{ID: AllPhotosAlbumID, Name: "All Photos"}}

	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		dir := filepath.Join(f.root, e.Name())
		has, err := dirHasImage(dir)
		if err != nil || !has {
			continue
		}
		albums = append(albums, Album{ID: e.Name(), Name: e.Name()})
	}

	sort.Slice(albums[1:], func(i, j int) bool {
		return albums[i+1].Name < albums[j+1].Name
	})
	return albums, nil
}

func (f *FilesystemSource) ListPhotos(ctx context.Context, albumID string) ([]Photo, error) {
	dir := f.root
	if albumID != AllPhotosAlbumID {
		dir = filepath.Join(f.root, albumID)
	}

	var photos []Photo
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if d.IsDir() {
			if d.Name() != filepath.Base(dir) && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !isImageFile(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(f.root, path)
		if err != nil {
			return err
		}
		photos = append(photos, Photo{ID: rel, Name: d.Name()})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing photos in %s: %w", dir, err)
	}

	sort.Slice(photos, func(i, j int) bool { return photos[i].ID < photos[j].ID })
	return photos, nil
}

func (f *FilesystemSource) LocalPath(_ context.Context, photoID string) (string, error) {
	return filepath.Join(f.root, photoID), nil
}

func isImageFile(name string) bool {
	return imageExts[strings.ToLower(filepath.Ext(name))]
}

// dirHasImage reports whether dir contains at least one image file,
// searching recursively but stopping at the first match.
func dirHasImage(dir string) (bool, error) {
	found := false
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found {
			return filepath.SkipAll
		}
		if d.IsDir() {
			if d.Name() != filepath.Base(dir) && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if isImageFile(d.Name()) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found, err
}

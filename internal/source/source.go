// Package source defines the adapter interface every photo source
// (local folder, network drive, and eventually iCloud/Google Photos)
// implements, so the TUI and renderer don't need to know where the
// bytes actually come from.
package source

import "context"

type Album struct {
	// ID is opaque to callers and only meaningful to the Source that issued it.
	ID   string
	Name string
}

type Photo struct {
	ID   string
	Name string
}

type Source interface {
	Name() string
	ListAlbums(ctx context.Context) ([]Album, error)
	ListPhotos(ctx context.Context, albumID string) ([]Photo, error)

	// LocalPath returns a filesystem path containing the photo's bytes,
	// fetching/caching first if the source isn't already local-disk backed.
	LocalPath(ctx context.Context, photoID string) (string, error)
}

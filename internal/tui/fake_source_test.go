package tui

import (
	"context"
	"fmt"

	"github.com/brandonapol/goodmorning/internal/source"
)

// fakeSource is an in-memory source.Source so model tests can drive full
// screen transitions deterministically, without touching the filesystem
// or a real terminal.
type fakeSource struct {
	name          string
	albums        []source.Album
	photosByAlbum map[string][]source.Photo
	listErr       error
}

func (f *fakeSource) Name() string { return f.name }

func (f *fakeSource) ListAlbums(context.Context) ([]source.Album, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.albums, nil
}

func (f *fakeSource) ListPhotos(_ context.Context, albumID string) ([]source.Photo, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.photosByAlbum[albumID], nil
}

func (f *fakeSource) LocalPath(_ context.Context, photoID string) (string, error) {
	return photoID, nil
}

var errFakeSource = fmt.Errorf("fake source failure")

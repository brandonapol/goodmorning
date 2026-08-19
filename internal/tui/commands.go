package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brandonapol/goodmorning/internal/render"
	"github.com/brandonapol/goodmorning/internal/source"
)

type albumsLoadedMsg struct {
	albums []source.Album
	err    error
}

type photosLoadedMsg struct {
	photos []source.Photo
	err    error
}

type artRenderedMsg struct {
	art   string
	width int
	err   error
}

func loadAlbums(src source.Source) tea.Cmd {
	return func() tea.Msg {
		albums, err := src.ListAlbums(context.Background())
		return albumsLoadedMsg{albums: albums, err: err}
	}
}

func loadPhotos(src source.Source, albumID string) tea.Cmd {
	return func() tea.Msg {
		photos, err := src.ListPhotos(context.Background(), albumID)
		return photosLoadedMsg{photos: photos, err: err}
	}
}

func renderPhoto(r *render.Renderer, src source.Source, photoID string, width int) tea.Cmd {
	return func() tea.Msg {
		path, err := src.LocalPath(context.Background(), photoID)
		if err != nil {
			return artRenderedMsg{err: err, width: width}
		}
		art, err := r.Render(path, width)
		return artRenderedMsg{art: art, width: width, err: err}
	}
}

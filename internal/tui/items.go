package tui

import "github.com/brandonapol/goodmorning/internal/source"

type sourceItem struct {
	src source.Source
}

func (i sourceItem) FilterValue() string { return i.src.Name() }
func (i sourceItem) Title() string       { return i.src.Name() }
func (i sourceItem) Description() string { return "" }

type albumItem struct {
	album source.Album
}

func (i albumItem) FilterValue() string { return i.album.Name }
func (i albumItem) Title() string       { return i.album.Name }
func (i albumItem) Description() string { return "" }

type photoItem struct {
	photo source.Photo
}

func (i photoItem) FilterValue() string { return i.photo.Name }
func (i photoItem) Title() string       { return i.photo.Name }
func (i photoItem) Description() string { return "" }

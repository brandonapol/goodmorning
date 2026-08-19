// Package tui implements the Bubble Tea application: pick a source, then
// an album, then a photo, then view it rendered as ANSI art full-screen.
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/brandonapol/goodmorning/internal/render"
	"github.com/brandonapol/goodmorning/internal/source"
)

type screen int

const (
	screenSources screen = iota
	screenAlbums
	screenPhotos
	screenDetail
)

var (
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Padding(0, 1)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)
)

type Model struct {
	sources  []source.Source
	renderer *render.Renderer

	screen screen

	sourceList list.Model
	albumList  list.Model
	photoList  list.Model

	activeSource source.Source
	activeAlbum  source.Album
	photos       []source.Photo
	photoIndex   int

	art     string
	artW    int // width the current art was rendered at
	loading bool
	err     error

	width, height int
}

func New(sources []source.Source) Model {
	m := Model{
		sources:  sources,
		renderer: render.NewRenderer(),
		screen:   screenSources,
	}

	srcItems := make([]list.Item, len(sources))
	for i, s := range sources {
		srcItems[i] = sourceItem{src: s}
	}
	m.sourceList = newList("Photo Sources", srcItems)
	m.albumList = newList("Albums", nil)
	m.photoList = newList("Photos", nil)

	return m
}

func newList(title string, items []list.Item) list.Model {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	return l
}

func (m Model) Init() tea.Cmd {
	if len(m.sources) == 1 {
		m.activeSource = m.sources[0]
		return loadAlbums(m.activeSource)
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleResize(msg)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.handleKey(msg)

	case albumsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		items := make([]list.Item, len(msg.albums))
		for i, a := range msg.albums {
			items[i] = albumItem{album: a}
		}
		m.albumList.SetItems(items)
		m.screen = screenAlbums
		return m, nil

	case photosLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.photos = msg.photos
		items := make([]list.Item, len(msg.photos))
		for i, p := range msg.photos {
			items[i] = photoItem{photo: p}
		}
		m.photoList.SetItems(items)
		m.screen = screenPhotos
		return m, nil

	case artRenderedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.art = msg.art
		m.artW = msg.width
		return m, nil
	}

	return m, nil
}

func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width, m.height = msg.Width, msg.Height

	listW, listH := msg.Width, msg.Height-2
	m.sourceList.SetSize(listW, listH)
	m.albumList.SetSize(listW, listH)
	m.photoList.SetSize(listW, listH)

	if m.screen == screenDetail && len(m.photos) > 0 {
		return m, m.renderCurrentPhoto()
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc", "backspace":
		return m.goBack()
	}

	switch m.screen {
	case screenSources:
		var cmd tea.Cmd
		m.sourceList, cmd = m.sourceList.Update(msg)
		if msg.String() == "enter" {
			if item, ok := m.sourceList.SelectedItem().(sourceItem); ok {
				m.activeSource = item.src
				m.loading = true
				m.err = nil
				return m, loadAlbums(m.activeSource)
			}
		}
		return m, cmd

	case screenAlbums:
		var cmd tea.Cmd
		m.albumList, cmd = m.albumList.Update(msg)
		if msg.String() == "enter" {
			if item, ok := m.albumList.SelectedItem().(albumItem); ok {
				m.activeAlbum = item.album
				m.loading = true
				m.err = nil
				return m, loadPhotos(m.activeSource, m.activeAlbum.ID)
			}
		}
		return m, cmd

	case screenPhotos:
		var cmd tea.Cmd
		m.photoList, cmd = m.photoList.Update(msg)
		if msg.String() == "enter" {
			m.photoIndex = m.photoList.Index()
			m.screen = screenDetail
			m.art = ""
			m.loading = true
			m.err = nil
			return m, m.renderCurrentPhoto()
		}
		return m, cmd

	case screenDetail:
		switch msg.String() {
		case "left", "h":
			return m.stepPhoto(-1)
		case "right", "l":
			return m.stepPhoto(1)
		}
	}

	return m, nil
}

func (m Model) stepPhoto(delta int) (tea.Model, tea.Cmd) {
	next := m.photoIndex + delta
	if next < 0 || next >= len(m.photos) {
		return m, nil
	}
	m.photoIndex = next
	m.art = ""
	m.loading = true
	m.err = nil
	return m, m.renderCurrentPhoto()
}

func (m Model) renderCurrentPhoto() tea.Cmd {
	if len(m.photos) == 0 {
		return nil
	}
	width := m.width
	if width <= 0 {
		width = 80
	}
	photo := m.photos[m.photoIndex]
	return renderPhoto(m.renderer, m.activeSource, photo.ID, width)
}

func (m Model) goBack() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenAlbums:
		if len(m.sources) > 1 {
			m.screen = screenSources
		}
	case screenPhotos:
		m.screen = screenAlbums
	case screenDetail:
		m.screen = screenPhotos
		m.art = ""
	}
	m.err = nil
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenSources:
		return m.sourceList.View()
	case screenAlbums:
		return m.withStatus(m.albumList.View())
	case screenPhotos:
		return m.withStatus(m.photoList.View())
	case screenDetail:
		return m.detailView()
	}
	return ""
}

func (m Model) withStatus(body string) string {
	if m.err != nil {
		return body + "\n" + errorStyle.Render("error: "+m.err.Error())
	}
	return body
}

func (m Model) detailView() string {
	if m.err != nil {
		return errorStyle.Render("error: "+m.err.Error()) + "\n" + helpStyle.Render("esc back")
	}
	if m.loading || m.art == "" {
		return statusStyle.Render("rendering...")
	}

	photo := m.photos[m.photoIndex]
	status := statusStyle.Render(fmt.Sprintf("%s  (%d/%d)", photo.Name, m.photoIndex+1, len(m.photos)))
	help := helpStyle.Render("← → prev/next   esc back   q quit")
	return m.art + "\n" + status + "  " + help
}

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brandonapol/goodmorning/internal/source"
)

func twoSourceModel() Model {
	m := New([]source.Source{
		&fakeSource{
			name:   "Alpha",
			albums: []source.Album{{ID: source.AllPhotosAlbumID, Name: "All Photos"}, {ID: "Trip", Name: "Trip"}},
			photosByAlbum: map[string][]source.Photo{
				"Trip": {{ID: "trip/a.jpg", Name: "a.jpg"}, {ID: "trip/b.jpg", Name: "b.jpg"}},
			},
		},
		&fakeSource{name: "Beta"},
	})
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	return sized.(Model)
}

func oneSourceModel() Model {
	m := New([]source.Source{
		&fakeSource{
			name:   "Solo",
			albums: []source.Album{{ID: source.AllPhotosAlbumID, Name: "All Photos"}},
			photosByAlbum: map[string][]source.Photo{
				source.AllPhotosAlbumID: {{ID: "a.jpg", Name: "a.jpg"}},
			},
		},
	})
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	return sized.(Model)
}

func enterKey() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }
func escKey() tea.KeyMsg   { return tea.KeyMsg{Type: tea.KeyEsc} }
func leftKey() tea.KeyMsg  { return tea.KeyMsg{Type: tea.KeyLeft} }
func rightKey() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRight} }
func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func mustCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a non-nil command")
	}
	return cmd()
}

func TestNew_BuildsSourceList(t *testing.T) {
	m := twoSourceModel()
	if m.screen != screenSources {
		t.Fatalf("screen = %v, want screenSources", m.screen)
	}
	if got := len(m.sourceList.Items()); got != 2 {
		t.Fatalf("sourceList has %d items, want 2", got)
	}
}

func TestInit_SingleSourceAutoLoadsAlbums(t *testing.T) {
	m := oneSourceModel()
	cmd := m.Init()
	msg := mustCmd(t, cmd)
	loaded, ok := msg.(albumsLoadedMsg)
	if !ok {
		t.Fatalf("Init() message type = %T, want albumsLoadedMsg", msg)
	}
	if len(loaded.albums) != 1 {
		t.Errorf("albums = %v, want 1 album", loaded.albums)
	}
}

func TestInit_MultiSourceWaitsForSelection(t *testing.T) {
	m := twoSourceModel()
	if cmd := m.Init(); cmd != nil {
		t.Error("Init() with multiple sources: want nil cmd, got non-nil")
	}
}

func TestSourceSelection_LoadsAlbums(t *testing.T) {
	m := twoSourceModel()

	updated, cmd := m.Update(enterKey())
	m = updated.(Model)
	if !m.loading {
		t.Error("after selecting a source: want loading=true")
	}

	msg := mustCmd(t, cmd)
	updated, _ = m.Update(msg)
	m = updated.(Model)

	if m.screen != screenAlbums {
		t.Fatalf("screen = %v, want screenAlbums", m.screen)
	}
	if m.loading {
		t.Error("after albums loaded: want loading=false")
	}
	if got := len(m.albumList.Items()); got != 2 {
		t.Errorf("albumList has %d items, want 2", got)
	}
}

func TestSourceSelection_ErrorSurfacesWithoutAdvancingScreen(t *testing.T) {
	m := New([]source.Source{
		&fakeSource{name: "Broken", listErr: errFakeSource},
		&fakeSource{name: "Other"},
	})
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = sized.(Model)

	updated, cmd := m.Update(enterKey())
	m = updated.(Model)
	msg := mustCmd(t, cmd)
	updated, _ = m.Update(msg)
	m = updated.(Model)

	if m.err == nil {
		t.Fatal("want m.err set after failed album load")
	}
	if m.screen != screenSources {
		t.Errorf("screen = %v, want screenSources to stay put on error", m.screen)
	}
}

func TestFullNavigationFlow(t *testing.T) {
	m := twoSourceModel()

	// source -> albums
	updated, cmd := m.Update(enterKey())
	m = updated.(Model)
	updated, _ = m.Update(mustCmd(t, cmd))
	m = updated.(Model)
	if m.screen != screenAlbums {
		t.Fatalf("screen = %v, want screenAlbums", m.screen)
	}

	// move selection onto "Trip" (index 1) and select it -> photos
	updated, _ = m.Update(runeKey('j'))
	m = updated.(Model)
	if got := m.albumList.Index(); got != 1 {
		t.Fatalf("albumList.Index() = %d, want 1 after moving down", got)
	}
	updated, cmd = m.Update(enterKey())
	m = updated.(Model)
	updated, _ = m.Update(mustCmd(t, cmd))
	m = updated.(Model)
	if m.screen != screenPhotos {
		t.Fatalf("screen = %v, want screenPhotos", m.screen)
	}
	if len(m.photos) != 2 {
		t.Fatalf("photos = %v, want 2", m.photos)
	}

	// select first photo -> detail
	updated, cmd = m.Update(enterKey())
	m = updated.(Model)
	if m.screen != screenDetail {
		t.Fatalf("screen = %v, want screenDetail immediately on selection", m.screen)
	}
	if !m.loading || m.art != "" {
		t.Errorf("entering detail: want loading=true and art empty, got loading=%v art=%q", m.loading, m.art)
	}
	if cmd == nil {
		t.Fatal("want a render command on entering detail")
	}

	// simulate successful render
	updated, _ = m.Update(artRenderedMsg{art: "ART", width: 100})
	m = updated.(Model)
	if m.art != "ART" || m.loading {
		t.Errorf("after artRenderedMsg: art=%q loading=%v, want art=ART loading=false", m.art, m.loading)
	}

	// step to next photo
	updated, cmd = m.Update(rightKey())
	m = updated.(Model)
	if m.photoIndex != 1 {
		t.Errorf("photoIndex = %d, want 1 after stepping right", m.photoIndex)
	}
	if !m.loading || m.art != "" {
		t.Error("after stepping to next photo: want loading=true and art cleared")
	}
	if cmd == nil {
		t.Error("want a render command after stepping to next photo")
	}

	// stepping right again should be a no-op (already at last photo)
	updated, cmd = m.Update(rightKey())
	m = updated.(Model)
	if m.photoIndex != 1 {
		t.Errorf("photoIndex = %d, want to stay at 1 past the last photo", m.photoIndex)
	}
	if cmd != nil {
		t.Error("stepping past the last photo: want nil cmd (no-op)")
	}

	// step back to first photo
	updated, cmd = m.Update(leftKey())
	m = updated.(Model)
	if m.photoIndex != 0 {
		t.Errorf("photoIndex = %d, want 0 after stepping left", m.photoIndex)
	}
	if cmd == nil {
		t.Error("want a render command after stepping to previous photo")
	}

	// stepping left again should be a no-op (already at first photo)
	updated, cmd = m.Update(leftKey())
	m = updated.(Model)
	if m.photoIndex != 0 {
		t.Errorf("photoIndex = %d, want to stay at 0 before the first photo", m.photoIndex)
	}
	if cmd != nil {
		t.Error("stepping before the first photo: want nil cmd (no-op)")
	}

	// esc: detail -> photos -> albums -> sources (two sources configured)
	updated, _ = m.Update(escKey())
	m = updated.(Model)
	if m.screen != screenPhotos {
		t.Fatalf("screen = %v, want screenPhotos after esc from detail", m.screen)
	}
	updated, _ = m.Update(escKey())
	m = updated.(Model)
	if m.screen != screenAlbums {
		t.Fatalf("screen = %v, want screenAlbums after esc from photos", m.screen)
	}
	updated, _ = m.Update(escKey())
	m = updated.(Model)
	if m.screen != screenSources {
		t.Fatalf("screen = %v, want screenSources after esc from albums (multi-source)", m.screen)
	}
}

func TestGoBack_SingleSourceStaysOnAlbums(t *testing.T) {
	m := oneSourceModel()
	updated, _ := m.Update(mustCmd(t, m.Init()))
	m = updated.(Model)
	if m.screen != screenAlbums {
		t.Fatalf("screen = %v, want screenAlbums", m.screen)
	}

	updated, _ = m.Update(escKey())
	m = updated.(Model)
	if m.screen != screenAlbums {
		t.Errorf("screen = %v, want to stay on screenAlbums (no source list to go back to)", m.screen)
	}
}

func TestQuit_FromAnyScreenAndCtrlC(t *testing.T) {
	m := twoSourceModel()

	_, cmd := m.Update(runeKey('q'))
	if _, ok := mustCmd(t, cmd).(tea.QuitMsg); !ok {
		t.Error("'q' should return a quit command")
	}

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if _, ok := mustCmd(t, cmd).(tea.QuitMsg); !ok {
		t.Error("ctrl+c should return a quit command")
	}
}

func TestWindowResize_ReRendersDetailScreen(t *testing.T) {
	m := twoSourceModel()
	updated, cmd := m.Update(enterKey())
	m = updated.(Model)
	updated, _ = m.Update(mustCmd(t, cmd))
	m = updated.(Model) // screenAlbums, All Photos selected by default

	updated, _ = m.Update(runeKey('j')) // move onto "Trip", which has photos
	m = updated.(Model)
	updated, cmd = m.Update(enterKey()) // select the Trip album
	m = updated.(Model)
	updated, _ = m.Update(mustCmd(t, cmd))
	m = updated.(Model) // screenPhotos

	updated, cmd = m.Update(enterKey()) // select first photo
	m = updated.(Model)
	updated, _ = m.Update(mustCmd(t, cmd))
	m = updated.(Model) // screenDetail with art loaded (fake render will error, that's fine)

	_, cmd = m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	if cmd == nil {
		t.Error("resizing while on the detail screen: want a re-render command")
	}
}

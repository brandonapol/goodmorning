package render

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// writeTestPNG writes a small solid-ish gradient PNG so real image decode
// and ASCII conversion code paths get exercised, not just plumbing.
func writeTestPNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for x := 0; x < 40; x++ {
		for y := 0; y < 30; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 6), G: uint8(y * 8), B: 128, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func newTestImage(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.png")
	writeTestPNG(t, path)
	return path
}

func writeGarbageFile(path string) error {
	return os.WriteFile(path, []byte("not an image"), 0o644)
}

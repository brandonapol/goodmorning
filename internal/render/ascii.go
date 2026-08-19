// Package render converts an image file on disk into terminal ANSI/ASCII
// art. The conversion itself is deterministic brightness/color sampling,
// not something that needs an LLM — TheZoraiz/ascii-image-converter does
// the heavy lifting; this package just wires it up with sizing and a
// disk cache so repeated views/resizes don't re-render from scratch.
package render

import (
	"fmt"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	"github.com/muesli/termenv"
)

type Renderer struct {
	cache   *diskCache // nil disables caching
	colored bool       // false on terminals with no color support (e.g. dumb/plain SSH)
}

// NewRenderer builds a Renderer with disk caching enabled and color output
// auto-detected from the terminal. If the cache directory can't be
// determined/created, caching is silently disabled rather than failing
// the whole app.
func NewRenderer() *Renderer {
	cache, _ := newDiskCache()
	return &Renderer{
		cache:   cache,
		colored: termenv.ColorProfile() != termenv.Ascii,
	}
}

// Render converts the image at path into ANSI-colored ASCII art sized to
// fit within the given terminal column width. Height is derived from the
// image's aspect ratio automatically.
func (r *Renderer) Render(path string, width int) (string, error) {
	if width <= 0 {
		width = 80
	}

	if r.cache != nil {
		if art, ok := r.cache.get(path, width, r.colored); ok {
			return art, nil
		}
	}

	flags := aic_package.DefaultFlags()
	flags.Colored = r.colored
	flags.Complex = true
	flags.Width = width

	art, err := aic_package.Convert(path, flags)
	if err != nil {
		return "", fmt.Errorf("rendering %s: %w", path, err)
	}

	if r.cache != nil {
		r.cache.put(path, width, r.colored, art)
	}
	return art, nil
}

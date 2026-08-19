package render

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// diskCache stores rendered ASCII/ANSI art on disk, keyed by the source
// file's identity (path + size + mtime) and the render width, so a resize
// or a new photo invalidates itself automatically without any bookkeeping.
type diskCache struct {
	dir string
}

func newDiskCache() (*diskCache, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(base, "goodmorning-photos", "ascii")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &diskCache{dir: dir}, nil
}

func (c *diskCache) key(path string, width int, colored bool) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%s|%d|%d|%d|%v", path, info.Size(), info.ModTime().UnixNano(), width, colored)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (c *diskCache) get(path string, width int, colored bool) (string, bool) {
	key, err := c.key(path, width, colored)
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(filepath.Join(c.dir, key))
	if err != nil {
		return "", false
	}
	return string(data), true
}

func (c *diskCache) put(path string, width int, colored bool, art string) {
	key, err := c.key(path, width, colored)
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(c.dir, key), []byte(art), 0o600)
}

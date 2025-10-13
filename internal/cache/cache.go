package cache

import (
	"os"
	"path/filepath"
)

const CacheDir = "internal/cache/files"

func GetCachedPath(originalPath, newExt string) string {
	base := filepath.Base(originalPath)
	name := base[:len(base)-len(filepath.Ext(base))] + newExt
	return filepath.Join(CacheDir, name)
}

func Save(path string, data []byte) error {
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, data, 0644)
}

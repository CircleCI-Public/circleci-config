package codebase

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalCodebase is a Codebase for files available on local disk
type LocalCodebase struct {
	BasePath string
}

func (c LocalCodebase) FindFileMatching(
	predicate func(string) bool,
	glob ...string,
) (string, error) {
	for _, g := range glob {
		filesFound, err := filepath.Glob(filepath.Join(c.BasePath, g))

		if err != nil {
			continue
		}

		for _, path := range filesFound {
			relPath, err := filepath.Rel(c.BasePath, path)
			if predicate(relPath) {
				return relPath, err
			}
		}
	}

	return "", fmt.Errorf("not found")
}

func (c LocalCodebase) FindFile(glob ...string) (path string, err error) {
	return c.FindFileMatching(func(string) bool { return true }, glob...)
}

func (c LocalCodebase) ReadFile(path string) (contents []byte, err error) {
	return os.ReadFile(filepath.Join(c.BasePath, path))
}

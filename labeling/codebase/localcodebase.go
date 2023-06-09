package codebase

import (
	"fmt"
	"io/fs"
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
) (foundPath string, err error) {
	basePath := c.BasePath
	if basePath == "" {
		basePath = "."
	}
	err = filepath.WalkDir(
		basePath,
		func(path string, d fs.DirEntry, fileError error) error {
			if fileError != nil {
				return fileError
			}
			relPath, innerErr := filepath.Rel(c.BasePath, path)
			if innerErr != nil {
				return innerErr
			}
			for _, g := range glob {
				matchesName, _ := filepath.Match(g, d.Name())
				matchesPath, _ := filepath.Match(g, relPath)
				if !(matchesName || matchesPath) {
					continue
				}
				if predicate(relPath) {
					foundPath = relPath
					return filepath.SkipAll
				}
			}
			return nil
		})

	if err != nil {
		return foundPath, err
	}

	if foundPath == "" {
		return foundPath, fmt.Errorf("not found")
	}

	return foundPath, nil
}

func (c LocalCodebase) FindFile(glob ...string) (path string, err error) {
	return c.FindFileMatching(func(string) bool { return true }, glob...)
}

func (c LocalCodebase) ReadFile(path string) (contents []byte, err error) {
	return os.ReadFile(filepath.Join(c.BasePath, path))
}

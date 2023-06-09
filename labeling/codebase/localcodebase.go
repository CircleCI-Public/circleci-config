package codebase

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const maxFiles = 250000

// LocalCodebase is a Codebase for files available on local disk
type LocalCodebase struct {
	BasePath string
	fileSet  []string
}

func (c LocalCodebase) FindFileMatching(
	predicate func(string) bool,
	glob ...string,
) (string, error) {
	files, err := c.files()
	if err != nil {
		return "", err
	}

	for _, path := range files {
		for _, g := range glob {
			matchesName, _ := filepath.Match(g, filepath.Base(path))
			matchesPath, _ := filepath.Match(g, path)
			if !(matchesName || matchesPath) {
				continue
			}
			if predicate(path) {
				return path, err
			}
		}
	}

	return "", fmt.Errorf("not found")

}

func (c LocalCodebase) files() ([]string, error) {
	if len(c.fileSet) > 0 {
		return c.fileSet, nil
	}

	basePath := c.BasePath
	if basePath == "" {
		basePath = "."
	}

	filesVisited := 0
	var fileList []string
	err := filepath.WalkDir(
		basePath,
		func(path string, d fs.DirEntry, fileError error) error {
			if fileError != nil {
				return fileError
			}
			relPath, innerErr := filepath.Rel(c.BasePath, path)
			if innerErr != nil {
				return innerErr
			}
			fileList = append(fileList, relPath)
			filesVisited++
			if filesVisited >= maxFiles {
				return filepath.SkipAll
			}
			return nil
		})

	c.fileSet = fileList

	return c.fileSet, err
}

func (c LocalCodebase) FindFile(glob ...string) (path string, err error) {
	return c.FindFileMatching(func(string) bool { return true }, glob...)
}

func (c LocalCodebase) ReadFile(path string) (contents []byte, err error) {
	return os.ReadFile(filepath.Join(c.BasePath, path))
}

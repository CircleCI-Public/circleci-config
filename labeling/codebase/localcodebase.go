package codebase

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxFiles = 250000
const defaultMaxDepth = 3

var NotFoundError = errors.New("not found")

// LocalCodebase is a Codebase for files available on local disk
type LocalCodebase struct {
	BasePath string
	maxDepth int // including root dir
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

	return "", NotFoundError

}

func (c LocalCodebase) files() ([]string, error) {
	if len(c.fileSet) > 0 {
		return c.fileSet, nil
	}

	basePath := c.BasePath
	if basePath == "" {
		basePath = "."
	}

	if c.maxDepth == 0 {
		c.maxDepth = defaultMaxDepth
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
			if d.IsDir() && pathDepth(relPath) > c.maxDepth {
				return filepath.SkipDir
			}
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

	sort.Slice(fileList, func(i, j int) bool {
		return pathDepth(fileList[i]) < pathDepth(fileList[j])
	})
	c.fileSet = fileList

	return c.fileSet, err
}

func pathDepth(path string) int {
	return strings.Count(path, string(os.PathSeparator)) + 1
}

func (c LocalCodebase) FindFile(glob ...string) (path string, err error) {
	return c.FindFileMatching(func(string) bool { return true }, glob...)
}

func (c LocalCodebase) ReadFile(path string) (contents []byte, err error) {
	return os.ReadFile(filepath.Join(c.BasePath, path))
}

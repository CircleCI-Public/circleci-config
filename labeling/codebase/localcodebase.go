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

func (c LocalCodebase) FindFile(glob string) (path string, err error) {
	filesFound, err := filepath.Glob(filepath.Join(c.BasePath, glob))
	if len(filesFound) < 1 {
		return path, fmt.Errorf("not found")
	}
	return filepath.Rel(c.BasePath, filesFound[0])
}

func (c LocalCodebase) ReadFile(path string) (contents []byte, err error) {
	return os.ReadFile(filepath.Join(c.BasePath, path))
}

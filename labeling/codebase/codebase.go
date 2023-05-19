package codebase

// Codebase interface that allows finding files in a codebase and reading file contents
type Codebase interface {
	// FindFile returns the path of the first file it finds matching `glob`
	// Note by go standard library doesn't understand globs with double asterisks,
	// like "/**/project.clj"
	FindFile(glob string) (path string, err error)
	ReadFile(path string) (contents []byte, err error)
}

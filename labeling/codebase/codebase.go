package codebase

// Codebase interface that allows finding files in a codebase and reading file contents
type Codebase interface {
	// FindFileMatching returns the path of the first file it finds matching a `glob`
	// and for which predicate returns true.
	// Note by go standard library doesn't understand globs with double asterisks,
	// like "/**/project.clj"
	FindFileMatching(predicate func(path string) bool, glob ...string) (path string, err error)
	// FindFile is like FindFileMatching, but with a constantly true predicate, so it always
	// returns the first file that matches glob
	FindFile(glob ...string) (path string, err error)
	ReadFile(path string) (contents []byte, err error)
}

package codebase

import (
	"path/filepath"
	"testing"
)

func TestLocalCodebase_FindFileMatching(t *testing.T) {
	tests := []struct {
		name         string
		BasePath     string
		predicate    func(string) bool
		globs        []string
		expectedPath string
		expectErr    bool
	}{
		{
			name:      "find.me found, but doesn't match predicate",
			predicate: func(s string) bool { return false },
			BasePath:  "./testdata",
			globs:     []string{"find.me"},
			expectErr: true,
		}, {
			name:         "*.go found, localcodebase.go matches predicate",
			predicate:    func(s string) bool { return s == "localcodebase.go" },
			globs:        []string{"*.go"},
			expectedPath: "localcodebase.go",
			expectErr:    false,
		}, {
			name:         "multiple globs",
			predicate:    func(s string) bool { return true },
			globs:        []string{"find.me", "testdata/find.me"},
			expectedPath: "testdata/find.me",
			expectErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := LocalCodebase{BasePath: tt.BasePath}
			gotPath, err := c.FindFileMatching(tt.predicate, tt.globs...)
			if (err != nil) != tt.expectErr {
				t.Errorf("FindFile() error %v, expectErr %v", err, tt.expectErr)
				return
			}
			if gotPath != tt.expectedPath {
				t.Errorf(" got %q, expected %q", gotPath, tt.expectedPath)
			}
		})
	}
}

func TestLocalCodebase_FindFile(t *testing.T) {
	tests := []struct {
		name         string
		BasePath     string
		globs        []string
		expectedPath string
		expectErr    bool
	}{
		{
			name:         "find.me found in testdata dir",
			BasePath:     "./testdata",
			globs:        []string{"find.me"},
			expectedPath: "find.me",
		}, {
			name:         "find.me found in current dir by glob",
			BasePath:     ".",
			globs:        []string{"find.me"},
			expectedPath: "testdata/find.me",
		}, {
			name:         "find.me found in current dir by extension",
			BasePath:     ".",
			globs:        []string{"*.me"},
			expectedPath: "testdata/find.me",
		}, {
			name:         "find.me found in current dir by exact path",
			BasePath:     ".",
			globs:        []string{"testdata/find.me"},
			expectedPath: "testdata/find.me",
		}, {
			name:         "find.me found in current dir, multiple globs, first matches",
			BasePath:     ".",
			globs:        []string{"find.me", "dontfind.me"},
			expectedPath: "testdata/find.me",
		}, {
			name:         "find.me found in current dir, multiple globs, last matches",
			BasePath:     ".",
			globs:        []string{"dontfind.me", "find.me"},
			expectedPath: "testdata/find.me",
		}, {
			name:         "localcodebase_test.go found in current dir",
			BasePath:     ".",
			globs:        []string{"localcodebase_test.go"},
			expectedPath: "localcodebase_test.go",
		}, {
			name:         "localcodebase_test.go found in current dir without BasePath",
			globs:        []string{"localcodebase_test.go"},
			expectedPath: "localcodebase_test.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := LocalCodebase{BasePath: tt.BasePath}
			gotPath, err := c.FindFile(tt.globs...)
			if (err != nil) != tt.expectErr {
				t.Errorf("FindFile() error %v, expectErr %v", err, tt.expectErr)
				return
			}
			if gotPath != tt.expectedPath {
				t.Errorf(" got %q, expected %q", gotPath, tt.expectedPath)
			}
		})
	}
}

func TestLocalCodebase_ReadFile(t *testing.T) {
	findMeExpectedContents := "test file for localcodebase tests"
	tests := []struct {
		name             string
		BasePath         string
		path             string
		expectedContents string
		expectErr        bool
	}{
		{
			name:             "can read find.me in testdata dir",
			BasePath:         "testdata",
			path:             "find.me",
			expectedContents: findMeExpectedContents,
			expectErr:        false,
		}, {
			name:             "can read find.me from current dir",
			path:             "testdata/find.me",
			expectedContents: findMeExpectedContents,
			expectErr:        false,
		}, {
			name:      "can't read non-existing file",
			path:      "cannot-find.me",
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := LocalCodebase{BasePath: tt.BasePath}
			gotContents, err := c.ReadFile(tt.path)
			if (err != nil) != tt.expectErr {
				t.Errorf("ReadFile() error %v, expectErr %v", err, tt.expectErr)
				return
			}
			gotString := string(gotContents)
			if gotString != tt.expectedContents {
				t.Errorf("got %q, expected %q", gotContents, tt.expectedContents)
			}
		})
	}
}

func TestLocalCodebase_files(t *testing.T) {
	t.Run("doesn't error", func(t *testing.T) {
		_, err := LocalCodebase{BasePath: "../.."}.files()
		if err != nil {
			t.Errorf("got %v, expected no error", err)
			return
		}
	})

	t.Run("README.md (at the root) comes before this file (localcodebase_test.go)", func(t *testing.T) {
		gotFiles, _ := LocalCodebase{BasePath: "../.."}.files()
		thisFilePos := -1
		readmePos := -1
		for i, f := range gotFiles {
			if f == "README.md" {
				readmePos = i
			} else if filepath.Base(f) == "localcodebase_test.go" {
				thisFilePos = i
			}
		}
		if thisFilePos == -1 {
			t.Error("this file not found")
		}
		if readmePos == -1 {
			t.Error("README.md file not found")
		}
		if readmePos > thisFilePos {
			t.Errorf("expected the order to be ascending but found files at positions (%d, %d)",
				readmePos,
				thisFilePos)
		}
	})

	t.Run("finds testdata contents", func(t *testing.T) {
		gotFiles, _ := LocalCodebase{BasePath: "../..", maxDepth: 3}.files()

		for _, f := range gotFiles {
			if filepath.Base(f) == "find.me" {
				// find.me file found
				return
			}
		}

		t.Error("find.me file not found")
	})

	t.Run("does not find testdata contents (too deep)", func(t *testing.T) {
		gotFiles, _ := LocalCodebase{BasePath: "../..", maxDepth: 2}.files()

		for _, f := range gotFiles {
			if filepath.Base(f) == "find.me" {
				t.Error("find.me file found, expected not to be found")
				return
			}
		}
		// find.me file not found
	})

}

func TestLocalCodebase_ListFiles(t *testing.T) {
	files, err := LocalCodebase{BasePath: "./testdata", maxDepth: 2}.ListFiles()

	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if len(files) != 1 {
		t.Errorf("unexpected files found, wanted 1 file but found %d", len(files))
	}

	if files[0] != "find.me" {
		t.Errorf("found unexpected file. \nwanted:\t\t%s\ngot:\t\t%s\n", "find.me", files[0])
	}

}

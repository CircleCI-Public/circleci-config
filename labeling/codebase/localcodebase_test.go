package codebase

import (
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

package codebase

import (
	"testing"
)

func TestLocalCodebase_FindFile(t *testing.T) {
	tests := []struct {
		name         string
		BasePath     string
		glob         string
		expectedPath string
		expectErr    bool
	}{
		{
			name:         "find.me found in testdata dir",
			BasePath:     "./testdata",
			glob:         "find.me",
			expectedPath: "find.me",
			expectErr:    false,
		}, {
			name:         "find.me found in current dir by glob",
			BasePath:     ".",
			glob:         "*/find.me",
			expectedPath: "testdata/find.me",
			expectErr:    false,
		}, {
			name:         "find.me found in current dir by exact path",
			BasePath:     ".",
			glob:         "testdata/find.me",
			expectedPath: "testdata/find.me",
			expectErr:    false,
		}, {
			name:         "localcodebase_test.go found in current dir",
			BasePath:     ".",
			glob:         "localcodebase_test.go",
			expectedPath: "localcodebase_test.go",
			expectErr:    false,
		}, {
			name:         "localcodebase_test.go found in current dir without BasePath",
			glob:         "localcodebase_test.go",
			expectedPath: "localcodebase_test.go",
			expectErr:    false,
		}, {
			name:      "find.me not found in current dir",
			BasePath:  ".",
			glob:      "find.me",
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := LocalCodebase{BasePath: tt.BasePath}
			gotPath, err := c.FindFile(tt.glob)
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

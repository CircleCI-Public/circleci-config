package labeling

import (
	"fmt"
	"github.com/CircleCI-Public/circleci-config/labeling/internal"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"path/filepath"
	"reflect"
	"testing"
)

// a codebase.Codebase for testing that just reads filenames and contents from a map
type fakeCodebase struct {
	// map of filename to file contents
	fileContents map[string]string
}

func (r fakeCodebase) FindFile(glob string) (path string, err error) {
	for k := range r.fileContents {
		matched, _ := filepath.Match(glob, k)
		if matched {
			return k, nil
		}
	}
	return "", fmt.Errorf("not found")
}

func (r fakeCodebase) ReadFile(path string) (contents []byte, err error) {
	contentString := r.fileContents[path]
	if contentString != "" {
		return []byte(contentString), nil
	}
	return nil, fmt.Errorf("not found")
}

func TestCodebase_ApplyAllRules(t *testing.T) {
	tests := []struct {
		name            string
		files           map[string]string
		expectedMatches []labels.Match
	}{
		{
			name: "go & node rules match",
			files: map[string]string{
				"go.mod":       "",
				"package.json": "{}",
			},
			expectedMatches: []labels.Match{
				{
					Label: labels.DepsNode,
					MatchData: labels.MatchData{
						BasePath:     ".",
						Dependencies: map[string]string{},
					},
				}, {
					Label: labels.DepsGo,
					MatchData: labels.MatchData{
						BasePath: ".",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCodebase{tt.files}
			expected := make(labels.MatchSet)
			for _, m := range tt.expectedMatches {
				// all should be Valid
				m.Valid = true
				expected[m.Label] = m
			}
			got := ApplyAllRules(c)

			if !reflect.DeepEqual(got, expected) {
				t.Errorf("\n"+
					"got      %v\n"+
					"expected %v",
					got,
					expected)
			}
		})
	}
}

func TestCodebase_ApplyRules_Node(t *testing.T) {
	rules := internal.NodeRules
	tests := []struct {
		name            string
		files           map[string]string
		rules           []labels.Rule
		expectedMatches []labels.Match
	}{
		{
			name: "deps:node with package.json in subdir",
			files: map[string]string{
				"project/package.json": `{}`,
			},
			expectedMatches: []labels.Match{
				{
					Label: labels.DepsNode,
					MatchData: labels.MatchData{
						BasePath:     "project",
						Dependencies: map[string]string{},
					},
				},
			},
		}, {
			name: "deps:node with dependencies",
			files: map[string]string{
				"package.json": `{"dependencies": {"mylib": ">3.0"}}`,
			},
			expectedMatches: []labels.Match{
				{
					Label: labels.DepsNode,
					MatchData: labels.MatchData{
						BasePath:     ".",
						Dependencies: map[string]string{"mylib": ">3.0"},
					},
				},
			},
		}, {
			name: "deps:node with dependencies & scripts",
			files: map[string]string{
				"package.json": `{"dependencies": {"mylib": ">3.0"}, "scripts": {"test": "echo ok"}}`,
			},
			expectedMatches: []labels.Match{
				{
					Label: labels.DepsNode,
					MatchData: labels.MatchData{
						BasePath:     ".",
						Dependencies: map[string]string{"mylib": ">3.0"},
						Tasks:        map[string]string{"test": "echo ok"},
					},
				},
			},
		}, {
			name: "deps:node and test:jest",
			files: map[string]string{
				"package.json": `{"dependencies": {"mylib": ">3.0"}, "devDependencies": {"jest": "1.0"}}`,
			},
			expectedMatches: []labels.Match{
				{
					Label: labels.DepsNode,
					MatchData: labels.MatchData{
						BasePath: ".",
						Dependencies: map[string]string{
							"mylib": ">3.0",
							"jest":  "1.0",
						},
					},
				}, {
					Label: labels.TestJest,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCodebase{tt.files}
			expected := make(labels.MatchSet)
			for _, m := range tt.expectedMatches {
				// all should be Valid
				m.Valid = true
				expected[m.Label] = m
			}
			got := ApplyRules(c, rules)

			if !reflect.DeepEqual(got, expected) {
				t.Errorf("\n"+
					"got      %v\n"+
					"expected %v",
					got,
					expected)
			}
		})
	}
}

func TestCodebase_ApplyRules_Go(t *testing.T) {
	rules := internal.GoRules
	tests := []struct {
		name            string
		files           map[string]string
		rules           []labels.Rule
		expectedMatches []labels.Match
	}{
		{
			name: "deps:go",
			files: map[string]string{
				"go.mod": "module mymod\n\ngo 1.18\n",
			},
			expectedMatches: []labels.Match{
				{
					Label: labels.DepsGo,
					MatchData: labels.MatchData{
						BasePath: ".",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCodebase{tt.files}
			expected := make(labels.MatchSet)
			for _, m := range tt.expectedMatches {
				// all should be Valid
				m.Valid = true
				expected[m.Label] = m
			}
			got := ApplyRules(c, rules)

			if !reflect.DeepEqual(got, expected) {
				t.Errorf("\n"+
					"got      %v\n"+
					"expected %v",
					got,
					expected)
			}
		})
	}
}

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
	contentsByPath map[string]string
}

func (c fakeCodebase) FindFileMatching(predicate func(string) bool, globs ...string) (string, error) {
	for _, g := range globs {
		for path := range c.contentsByPath {
			matched, _ := filepath.Match(g, path)
			if matched && predicate(path) {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("not found")
}

func (c fakeCodebase) FindFile(globs ...string) (path string, err error) {
	return c.FindFileMatching(func(string) bool { return true }, globs...)
}

func (c fakeCodebase) ReadFile(path string) (contents []byte, err error) {
	contentString := c.contentsByPath[path]
	if contentString != "" {
		return []byte(contentString), nil
	}
	return nil, fmt.Errorf("not found")
}

func TestCodebase_ApplyAllRules(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]string
		expectedLabels []labels.Label
	}{
		{
			name: "go & node rules apply",
			files: map[string]string{
				"go.mod":       "",
				"package.json": "{}",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsNode,
					LabelData: labels.LabelData{
						BasePath:     ".",
						Dependencies: map[string]string{},
					},
				}, {
					Key: labels.DepsGo,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCodebase{tt.files}
			expected := make(labels.LabelSet)
			for _, label := range tt.expectedLabels {
				// all should be Valid
				label.Valid = true
				expected[label.Key] = label
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
		name           string
		files          map[string]string
		rules          []labels.Rule
		expectedLabels []labels.Label
	}{
		{
			name: "deps:node with package.json in subdir",
			files: map[string]string{
				"project/package.json": `{}`,
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsNode,
					LabelData: labels.LabelData{
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
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsNode,
					LabelData: labels.LabelData{
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
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsNode,
					LabelData: labels.LabelData{
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
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsNode,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"mylib": ">3.0",
							"jest":  "1.0",
						},
					},
				}, {
					Key: labels.TestJest,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCodebase{tt.files}
			expected := make(labels.LabelSet)
			for _, label := range tt.expectedLabels {
				// all should be Valid
				label.Valid = true
				expected[label.Key] = label
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
		name           string
		files          map[string]string
		rules          []labels.Rule
		expectedLabels []labels.Label
	}{
		{
			name: "deps:go",
			files: map[string]string{
				"go.mod": "module mymod\n\ngo 1.18\n",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsGo,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCodebase{tt.files}
			expected := make(labels.LabelSet)
			for _, label := range tt.expectedLabels {
				// all should be Valid
				label.Valid = true
				expected[label.Key] = label
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

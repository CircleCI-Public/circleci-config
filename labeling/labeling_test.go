package labeling

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/internal"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

// a codebase.Codebase for testing that just reads filenames and contents from a map
type fakeCodebase struct {
	// map of filename to file contents
	contentsByPath map[string]string
}

func (c fakeCodebase) FindFileMatching(predicate func(string) bool, globs ...string) (string, error) {
	for _, g := range globs {
		for path := range c.contentsByPath {
			matchesName, _ := filepath.Match(g, filepath.Base(path))
			matchesPath, _ := filepath.Match(g, path)
			if (matchesName || matchesPath) && predicate(path) {
				return path, nil
			}
		}
	}
	return "", codebase.NotFoundError
}

func (c fakeCodebase) FindFile(globs ...string) (path string, err error) {
	return c.FindFileMatching(func(string) bool { return true }, globs...)
}

func (c fakeCodebase) ReadFile(path string) (contents []byte, err error) {
	contentString := c.contentsByPath[path]
	if contentString != "" {
		return []byte(contentString), nil
	}
	return nil, codebase.NotFoundError
}

func TestCodebase_ApplyAllRules(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]string
		expectedLabels []labels.Label
	}{
		{
			name: "All rules apply",
			files: map[string]string{
				"go.mod":              "",
				"package.json":        `{"devDependencies":{"jest": "version"}}`,
				"cmd/cmd.go":          "package main",
				"rust-dir/Cargo.toml": "",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsNode,
					LabelData: labels.LabelData{
						BasePath:     ".",
						Dependencies: map[string]string{"jest": "version"},
					},
				}, {
					Key:       labels.TestJest,
					LabelData: labels.LabelData{},
				}, {
					Key: labels.DepsGo,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				}, {
					Key:       labels.ArtifactGoExecutable,
					LabelData: labels.LabelData{},
				}, {
					Key: labels.DepsRust,
					LabelData: labels.LabelData{
						BasePath: "rust-dir",
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
		{
			name: "deps:node, scripts and package_manager:yarn",
			files: map[string]string{
				"package.json": `{"dependencies": {"mylib": ">3.0"}, "devDependencies": {"jest": "1.0"}}`,
				"yarn.lock":    "yarn.lock file",
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
						HasLockFile: true,
					},
				}, {
					Key: labels.PackageManagerYarn,
				},
				{
					Key: labels.TestJest,
				},
			},
		},
		{
			name: "deps:node without any lock file",
			files: map[string]string{
				"package.json": `{"dependencies": {"mylib": ">3.0"}}`,
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsNode,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"mylib": ">3.0",
						},
						HasLockFile: false,
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

func TestCodebase_ApplyRules_Go(t *testing.T) {
	rules := internal.GoRules
	tests := []struct {
		name           string
		files          map[string]string
		rules          []labels.Rule
		expectedLabels []labels.Label
	}{
		{
			name: "go.mod => deps:go",
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
		}, {
			name: "go.mod in subdir => deps.go",
			files: map[string]string{
				"x/go.mod": "module mymod\n\ngo 1.18\n",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsGo,
					LabelData: labels.LabelData{
						BasePath: "x",
					},
				},
			},
		}, {
			name: "go.mod & go main package => artifact:go-executable",
			files: map[string]string{
				"go.mod":  "module mymod\n\ngo 1.18\n",
				"main.go": "package main",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsGo,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				}, {
					Key: labels.ArtifactGoExecutable,
				},
			},
		}, {
			name: "go.mod & go main package in subdir => artifact:go-executable",
			files: map[string]string{
				"go.mod":      "module mymod\n\ngo 1.18\n",
				"pkg/main.go": "package main",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsGo,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				}, {
					Key: labels.ArtifactGoExecutable,
				},
			},
		}, {
			name: "go.mod & go main package in subdir of cmd => artifact:go-executable",
			files: map[string]string{
				"go.mod":        "module mymod\n\ngo 1.18\n",
				"cmd/x/main.go": "//this is package main\npackage main",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsGo,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				}, {
					Key: labels.ArtifactGoExecutable,
				},
			},
		}, {
			name: "go main package without go.mod => no labels",
			files: map[string]string{
				"main.go": "package main",
			},
			expectedLabels: []labels.Label{},
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

func TestCodebase_ApplyRules_Python(t *testing.T) {
	rules := internal.PythonRules
	tests := []struct {
		name           string
		files          map[string]string
		rules          []labels.Rule
		expectedLabels []labels.Label
	}{
		{
			name: "requirements.txt => deps:python",
			files: map[string]string{
				"requirements.txt": "mylib==1.0",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
			},
		}, {
			name: "requirements.txt in subdir => deps:python",
			files: map[string]string{
				"x/requirements.txt": "mylib==1.0",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: "x",
					},
					Valid: true,
				},
			},
		}, {
			name: "Pipfile => package_manager:pipenv",
			files: map[string]string{
				"Pipfile": "[packages]\nmylib = \"==1.0\"\n",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
				{
					Key: labels.PackageManagerPipenv,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
			},
		},
		{
			name: "poetry.lock => package_manager:poetry",
			files: map[string]string{
				"poetry.lock": "mylib==1.0",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
				{
					Key: labels.PackageManagerPoetry,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
			},
		},
		{
			name: "manage.py => file:manage.py",
			files: map[string]string{
				"manage.py": "mylib==1.0",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
				{
					Key: labels.FileManagePy,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
			},
		},
		{
			name: "pyproject.toml contains pipenv => package_manager:pipenv",
			files: map[string]string{
				"pyproject.toml": "[tool.pipenv]\nname = \"mylib\"\n",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
				{
					Key: labels.PackageManagerPipenv,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
			},
		},
		{
			name: "pyproject.toml contains poetry => package_manager:poetry",
			files: map[string]string{
				"pyproject.toml": "[tool.poetry]\nname = \"mylib\"\n",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
				{
					Key: labels.PackageManagerPoetry,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
					Valid: true,
				},
			},
		},
		{
			name: "pyproject.toml in subdir contains poetry => package_manager:poetry",
			files: map[string]string{
				"x/pyproject.toml": "[tool.poetry]\nname = \"mylib\"\n",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: "x",
					},
					Valid: true,
				},
				{
					Key: labels.PackageManagerPoetry,
					LabelData: labels.LabelData{
						BasePath: "x",
					},
					Valid: true,
				},
			},
		},
		{
			name: "project contains .python-version => python:3.7 dependency",
			files: map[string]string{
				"requirements.txt": "mylib==1.0",
				".python-version":  "3.7\n",
			},
			expectedLabels: []labels.Label{
				{
					Key: labels.DepsPython,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"python": "3.7",
						},
					},
					Valid: true,
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

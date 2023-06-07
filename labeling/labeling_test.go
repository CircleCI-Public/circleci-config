package labeling

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

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
			name: "All rules apply",
			files: map[string]string{
				"go.mod":       "",
				"package.json": `{"devDependencies":{"jest": "version"}}`,
				"cmd/cmd.go":   "package main",
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
				},
			},
		},
		{
			name: "Ruby version",
			files: map[string]string{
				"Gemfile": rubyGemfile,
			},

			expectedLabels: []labels.Label{
				{
					Key: labels.DepsRuby,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"ruby": "2.7.8",
						},
					},
				},
			},
		},
		{
			name: "Ruby version w/ rspec",
			files: map[string]string{
				"Gemfile": rubyGemfileWithRailsRSpec,
			},

			expectedLabels: []labels.Label{
				{
					Key: labels.DepsRuby,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"ruby":  "2.7.8",
							"rspec": "true",
						},
					},
				},
			},
		},
		{
			name: "Ruby version with engine info",
			files: map[string]string{
				"Gemfile": rubyGemfileWithEngine,
			},

			expectedLabels: []labels.Label{
				{
					Key: labels.DepsRuby,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"ruby": "1.9.3",
						},
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
					},
				}, {
					Key: labels.PackageManagerYarn,
				},
				{
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

const rubyGemfile = `
source 'https://rubygems.org'

git_source(:github) do |repo_name|
  repo_name = "#{repo_name}/#{repo_name}" unless repo_name.include?('/')
  "https://github.com/#{repo_name}.git"
end

ruby '2.7.8'

# Bundle edge Rails instead: gem 'rails', github: 'rails/rails'
gem 'rails', '~> 6.0.1'
`

const rubyGemfileWithRailsRSpec = `
source 'https://rubygems.org'

git_source(:github) do |repo_name|
  repo_name = "#{repo_name}/#{repo_name}" unless repo_name.include?('/')
  "https://github.com/#{repo_name}.git"
end

ruby '2.7.8'

# Bundle edge Rails instead: gem 'rails', github: 'rails/rails'
gem 'rails', '~> 6.0.1'

gem 'rspec-rails', '4.0.0.beta3'

`

const rubyGemfileWithEngine = `
source 'https://rubygems.org'

git_source(:github) do |repo_name|
  repo_name = "#{repo_name}/#{repo_name}" unless repo_name.include?('/')
  "https://github.com/#{repo_name}.git"
end

ruby '1.9.3', :engine => 'jruby', :engine_version => '1.6.7'

# Bundle edge Rails instead: gem 'rails', github: 'rails/rails'
gem 'rails', '~> 6.0.1'`

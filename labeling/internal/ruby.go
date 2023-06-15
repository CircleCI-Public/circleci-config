package internal

import (
	"errors"
	"path"
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var RubyRules = []labels.Rule{
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.DepsRuby
		gemspecPath, err := c.FindFile("*.gemspec")
		if err != nil && !errors.Is(err, codebase.NotFoundError) {
			return label, err
		}
		if gemspecPath != "" {
			label.Valid = true
			label.BasePath = path.Dir(gemspecPath)
			return readDepsFile(c, label, gemspecPath)
		}

		gemfilePath, err := c.FindFile("Gemfile")
		if err != nil {
			return label, err
		}
		label.Valid = gemfilePath != ""
		label.BasePath = path.Dir(gemfilePath)
		return readDepsFile(c, label, gemfilePath)
	},
}

// Parse the Gemfile to add dependencies to the label
func readDepsFile(c codebase.Codebase, label labels.Label, filePath string) (labels.Label, error) {
	fileContents, err := c.ReadFile(filePath)
	if err != nil {
		return label, err
	}
	label.Dependencies = make(map[string]string)

	for _, line := range strings.Split(string(fileContents), "\n") {
		if strings.HasPrefix(line, "ruby ") {
			version := strings.Split(line, ",")[0]
			version = strings.SplitAfter(version, "ruby ")[1]
			version = strings.ReplaceAll(version, "'", "")
			label.Dependencies["ruby"] = version
		}

		if strings.Contains(line, "gem 'rspec-rails'") {
			label.Dependencies["rspec"] = "true"
		}

		if strings.Contains(line, "development_dependency('rake'") {
			label.Dependencies["rake"] = "true"
		}

		if strings.Contains(line, "gem 'pg'") {
			label.Dependencies["pg"] = "true"
		}
	}
	return label, nil
}

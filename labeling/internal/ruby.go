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
		label.Dependencies = make(map[string]string)

		gemfilePath, err := c.FindFile("Gemfile")
		if err != nil && !errors.Is(err, codebase.NotFoundError) {
			return label, err
		}
		if gemfilePath != "" {
			label.Valid = true
			label.BasePath = path.Dir(gemfilePath)
			err = readDepsFile(c, label.Dependencies, gemfilePath)
			if err != nil {
				return label, err
			}
			label.HasLockFile = hasPath(c, path.Join(path.Dir(gemfilePath), "Gemfile.lock"))
		}
		return label, nil
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.PackageManagerGemspec
		label.Dependencies = make(map[string]string)

		gemspecPath, err := c.FindFile("*.gemspec")
		if err != nil && !errors.Is(err, codebase.NotFoundError) {
			return label, err
		}

		if gemspecPath != "" {
			label.Valid = true
			label.BasePath = path.Dir(gemspecPath)
			err = readDepsFile(c, label.Dependencies, gemspecPath)
			if err != nil {
				return label, err
			}
		}
		return label, nil
	},
}

// Parse the Gemfile to add dependencies to the label
func readDepsFile(c codebase.Codebase, deps map[string]string, filePath string) error {
	fileContents, err := c.ReadFile(filePath)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(fileContents), "\n") {
		line = strings.ReplaceAll(line, "\"", "'")
		if strings.HasPrefix(line, "ruby ") {
			version := strings.Split(line, ",")[0]
			version = strings.SplitAfter(version, "ruby ")[1]
			version = strings.ReplaceAll(version, "'", "")
			version = strings.TrimSpace(version)
			deps["ruby"] = version
		}

		if strings.Contains(line, "gem 'rspec-rails'") ||
			strings.Contains(line, "gem 'rspec'") {
			deps["rspec"] = "true"
		}

		if strings.Contains(line, "gem 'rspec_junit_formatter'") ||
			strings.Contains(line, "development_dependency('rspec_junit_formatter'") {
			deps["rspec_junit_formatter"] = "true"
		}

		if strings.Contains(line, "development_dependency('rake'") ||
			strings.Contains(line, "gem 'rake'") {
			deps["rake"] = "true"
		}

		if strings.Contains(line, "gem 'pg'") {
			deps["pg"] = "true"
		}
	}
	return nil
}

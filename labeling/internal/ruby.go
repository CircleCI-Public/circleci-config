package internal

import (
	"errors"
	"path"
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var RubyRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
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
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
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

		gems := map[string]string{
			"rails":                 "rails",
			"rspec-rails":           "rspec",
			"rspec":                 "rspec",
			"rspec_junit_formatter": "rspec_junit_formatter",
			"rake":                  "rake",
			"pg":                    "pg",
		}
		for k, v := range gems {
			if hasGem(line, k) {
				deps[v] = "true"
			}
		}
	}
	return nil
}

func hasGem(line string, gem string) bool {
	forms := []string{"gem '", "development_dependency '", "development_dependency('"}
	for _, f := range forms {
		if strings.Contains(line, f+gem) {
			return true
		}
	}
	return false
}

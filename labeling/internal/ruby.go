package internal

import (
	"path"
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var RubyRules = []labels.Rule{
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.DepsRuby
		gemfilePath, err := c.FindFile("Gemfile", "*/Gemfile")
		label.Valid = gemfilePath != ""
		label.BasePath = path.Dir(gemfilePath)
		return readGemfile(c, label, gemfilePath)
	},
}

// Parse the Gemfile to add dependencies to the label
func readGemfile(c codebase.Codebase, label labels.Label, filePath string) (labels.Label, error) {
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
	}
	return label, nil
}

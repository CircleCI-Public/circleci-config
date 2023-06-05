package internal

import (
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var RubyRules = []labels.Rule{
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.PackageManagerBundler
		gemfilePath, err := c.FindFile("Gemfile", "*/Gemfile")
		label.Valid = gemfilePath != ""
		label.BasePath = path.Dir(gemfilePath)
		return label, err
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.TestRSpec
		rspecConfigPath, err := c.FindFile(".rspec")
		label.Valid = rspecConfigPath != ""
		return label, err
	},
}

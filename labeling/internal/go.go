package internal

import (
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"path"
)

var GoRules = []labels.Rule{
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.DepsGo
		goModPath, err := c.FindFile("go.mod")
		label.Valid = goModPath != ""
		label.BasePath = path.Dir(goModPath)
		return label, err
	},
}

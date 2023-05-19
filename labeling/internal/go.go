package internal

import (
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"path"
)

var GoRules = []labels.Rule{
	func(c codebase.Codebase, ms *labels.MatchSet) (m labels.Match, err error) {
		m.Label = labels.DepsGo
		goModPath, err := c.FindFile("go.mod")
		m.Valid = goModPath != ""
		m.BasePath = path.Dir(goModPath)
		return m, err
	},
}

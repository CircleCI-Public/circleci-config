package internal

import (
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var GithubActionRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.CICDGithubActions
		configPath, err := c.FindFile(".github/workflows/*.yml", ".github/workflows/*.yaml")
		label.Valid = configPath != ""
		label.BasePath = path.Dir(configPath)
		return label, err
	},
}

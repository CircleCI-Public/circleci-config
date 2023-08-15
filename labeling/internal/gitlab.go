package internal

import (
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var GitlabWorkflowRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.CICDGitlabWorkflow
		configPath, err := c.FindFile(".gitlab-ci.yml")
		label.Valid = configPath != ""
		label.BasePath = path.Dir(configPath)
		return label, err
	},
}

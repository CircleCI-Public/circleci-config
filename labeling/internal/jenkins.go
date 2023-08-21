package internal

import (
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var JenkinsRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.CICDJenkins
		configPath, err := c.FindFile("Jenkinsfile")
		label.Valid = configPath != ""
		label.BasePath = path.Dir(configPath)
		return label, err
	},
}

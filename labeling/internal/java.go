package internal

import (
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var JavaRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.DepsJava
		pomXml, err := c.FindFile("pom.xml", "gradlew")
		label.Valid = pomXml != ""
		label.BasePath = path.Dir(pomXml)
		return label, err
	},
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.ToolGradle
		gradlew, err := c.FindFile("gradlew")
		label.Valid = gradlew != "" && path.Dir(gradlew) == ls[labels.DepsJava].BasePath
		return label, err
	},
}

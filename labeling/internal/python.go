package internal

import (
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var pipenvFiles = []string{
	"Pipfile",
	"Pipfile.lock",
}

var poetryFiles = []string{
	"pyproject.toml",
	"poetry.lock",
}

// All the possible files that could be used to determine if it's a Python codebase
var possiblePythonFiles = append(
	append(
		[]string{
			"requirements.txt",
		},
		pipenvFiles...,
	),
	poetryFiles...,
)

var PythonRules = []labels.Rule{
	func(c codebase.Codebase, ls *labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key: labels.DepsPython,
		}
		filePath, _ := c.FindFile(possiblePythonFiles...)
		label.Valid = filePath != ""
		label.BasePath = path.Dir(filePath)
		return label, nil
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key: labels.PackageManagerPipenv,
		}
		pipfile, _ := c.FindFile(pipenvFiles...)
		label.Valid = pipfile != ""
		label.BasePath = path.Dir(pipfile)
		return label, nil
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key: labels.PackageManagerPoetry,
		}
		poetryLock, _ := c.FindFile(poetryFiles...)
		label.Valid = poetryLock != ""
		label.BasePath = path.Dir(poetryLock)
		return label, nil
	},
}

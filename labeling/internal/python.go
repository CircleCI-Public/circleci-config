package internal

import (
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var possiblePythonFiles = []string{
	"requirements.txt",
	"*/requirements.txt",
	"Pipfile",
	"*/Pipfile",
	"Pipfile.lock",
	"*/Pipfile.lock",
	"setup.py",
	"*/setup.py",
	"poetry.lock",
}

var PythonRules = []labels.Rule{
	func(c codebase.Codebase, ls *labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key:   labels.DepsPython,
			Valid: false,
		}
		filePath, _ := c.FindFile(possiblePythonFiles...)
		label.Valid = filePath != ""
		label.BasePath = path.Dir(filePath)
		return label, nil
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key:   labels.PackageManagerPipenv,
			Valid: false,
		}
		pipfile, _ := c.FindFile("Pipfile", "*/Pipfile", "Pipfile.lock", "*/Pipfile.lock")
		label.Valid = pipfile != ""
		label.BasePath = path.Dir(pipfile)
		return label, nil
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key:   labels.PackageManagerPoetry,
			Valid: false,
		}
		poetryLock, _ := c.FindFile("poetry.lock")
		label.Valid = poetryLock != ""
		label.BasePath = path.Dir(poetryLock)
		return label, nil
	},
}

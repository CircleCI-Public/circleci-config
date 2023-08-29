package internal

import (
	"path"
	"regexp"
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var pipenvFiles = []string{
	"Pipfile",
	"Pipfile.lock",
}

var poetryFiles = []string{
	"poetry.lock",
}

// All the possible files that could be used to determine if it's a Python codebase
var possiblePythonFiles = append(
	append(
		[]string{
			"requirements.txt",
			"pyproject.toml",
			"manage.py",
		},
		pipenvFiles...,
	),
	poetryFiles...,
)

var PythonRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key: labels.DepsPython,
		}
		filePath, _ := c.FindFile(possiblePythonFiles...)
		label.Valid = filePath != ""
		label.BasePath = path.Dir(filePath)

		pythonVersion := getPythonVersion(c)
		if pythonVersion != "" {
			label.Dependencies = map[string]string{
				"python": getPythonVersion(c),
			}
		}

		return label, nil
	},
	func(c codebase.Codebase, ls labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key: labels.PackageManagerPipenv,
		}
		pipfile, _ := c.FindFile(pipenvFiles...)
		label.Valid = pipfile != ""
		label.BasePath = path.Dir(pipfile)

		pyprojectPath, _ := c.FindFile("pyproject.toml")
		if pyprojectPath != "" && fileContainsString(c, pyprojectPath, "pipenv") {
			label.Valid = true
			label.BasePath = path.Dir(pyprojectPath)
		}

		return label, nil
	},
	func(c codebase.Codebase, ls labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key: labels.PackageManagerPoetry,
		}
		poetryLock, _ := c.FindFile(poetryFiles...)
		label.Valid = poetryLock != ""
		label.BasePath = path.Dir(poetryLock)

		pyprojectPath, _ := c.FindFile("pyproject.toml")
		if pyprojectPath != "" && fileContainsString(c, pyprojectPath, "poetry") {
			label.Valid = true
			label.BasePath = path.Dir(pyprojectPath)
		}

		return label, nil
	},
	func(c codebase.Codebase, ls labels.LabelSet) (labels.Label, error) {
		label := labels.Label{
			Key: labels.FileManagePy,
		}
		managePyPath, _ := c.FindFile("manage.py")
		label.Valid = managePyPath != ""
		label.BasePath = path.Dir(managePyPath)
		return label, nil
	},
}

func fileContainsString(c codebase.Codebase, filePath string, str string) bool {
	file, err := c.ReadFile(filePath)
	if err != nil {
		return false
	}

	fileStr := string(file)
	if fileStr == "" {
		return false
	}

	return strings.Contains(fileStr, str)
}

func getPythonVersion(c codebase.Codebase) string {
	versionRegex := regexp.MustCompile(`[0-9.]+`)

	versionFilePath, _ := c.FindFile(".python-version")
	if versionFilePath == "" {
		return ""
	}

	file, err := c.ReadFile(versionFilePath)
	if err != nil {
		return ""
	}

	version := versionRegex.FindString(string(file))
	if version != "" {
		return version
	}

	return ""
}

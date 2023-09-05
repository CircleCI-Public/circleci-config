package internal

import (
	"log"
	"path"
	"regexp"
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"github.com/pelletier/go-toml"
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
				"python": pythonVersion,
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

	versionFilePath, _ := c.FindFile(".python-version")
	if versionFilePath != "" {
		file, err := c.ReadFile(versionFilePath)
		if err != nil {
			log.Println("Unable to read file .python-version. Attempting to determine python version another way... ", err)
		} else {
			versionRegex := regexp.MustCompile(`[0-9.]+`)
			pythonVersion := versionRegex.FindString(string(file))
			if pythonVersion != "" {
				log.Println("Unable to parse file .python-version. Attempting to determine python version another way...")

			}
			log.Println("Found Python version in .python-version file: ", pythonVersion)
			return pythonVersion
		}
	}

	pyprojectFilePath, _ := c.FindFile("pyproject.toml")
	if pyprojectFilePath != "" {
		file, err := c.ReadFile(pyprojectFilePath)
		if err != nil {
			log.Println("Unable to read file pyproject.toml. Attempting to determine python version another way... ", err)
		} else {
			tree, err := toml.LoadBytes(file)
			if err != nil {
				log.Println("Unable to read file pyproject.toml. Attempting to determine python version another way... ", err)
			} else {
				pythonVersion := tree.Get("tool.poetry.dependencies.python")
				if pythonVersion == nil || pythonVersion.(string) == "" {
					log.Println("Error parsing pyproject.toml, python version is nil")
				} else {
					pythonVersion = strings.TrimPrefix(pythonVersion.(string), "^")
					log.Println("Found Python version in pyproject.toml file: ", pythonVersion.(string))
					return pythonVersion.(string)
				}
			}
		}
	}

	pipfileFilePath, _ := c.FindFile("Pipfile")
	if pipfileFilePath != "" {
		file, err := c.ReadFile(pipfileFilePath)
		if err != nil {
			log.Println("Unable to read file Pipfile. Attempting to determine python version another way... ", err)
		} else {
			tree, err := toml.LoadBytes(file)
			if err != nil {
				log.Println("Unable to read file Pipfile. Attempting to determine python version another way... ", err)
			} else {
				pythonVersion := tree.Get("requires.python_version")
				if pythonVersion == nil || pythonVersion.(string) == "" {
					log.Println("Error parsing Pipfile, returning nil")
					return ""
				} else {
					pythonVersion = strings.TrimPrefix(pythonVersion.(string), "^")
					log.Println("Found Python version in Pipfile: ", pythonVersion.(string))
					return pythonVersion.(string)
				}
			}
		}
	}
	return ""
}

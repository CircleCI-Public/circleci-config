package internal

import (
	"encoding/json"
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var NodeRules = []labels.Rule{
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.DepsNode
		packagePath := findPackageJSON(c)
		label.Valid = packagePath != ""
		if !label.Valid {
			return label, err
		}
		err = readPackageJSON(c, packagePath, &label)
		return label, err
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.PackageManagerYarn
		yarnLock, _ := c.FindFile("yarn.lock")
		label.Valid = yarnLock != ""
		return label, err
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.TestJest
		label.Valid = hasDependency(ls, "jest")
		return label, err
	},
}

// npmPackageJSON for unmarshalling npm package.json files
type npmPackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

func findPackageJSON(c codebase.Codebase) string {
	file, _ := c.FindFile("package.json", "*/package.json")
	return file
}

func readPackageJSON(c codebase.Codebase, filePath string, label *labels.Label) error {
	var packageJSON npmPackageJSON

	fileContents, err := c.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileContents, &packageJSON)
	if err != nil {
		return err
	}

	label.LabelData.BasePath = path.Dir(filePath)
	label.Tasks = packageJSON.Scripts
	label.Dependencies = make(map[string]string)

	for k, v := range packageJSON.Dependencies {
		label.Dependencies[k] = v
	}
	for k, v := range packageJSON.DevDependencies {
		label.Dependencies[k] = v
	}

	return err
}

func hasDependency(ls *labels.LabelSet, dep string) bool {
	return (*ls)[labels.DepsNode].Dependencies[dep] != ""
}

package internal

import (
	"encoding/json"
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"path"
)

var NodeRules = []labels.Rule{
	func(c codebase.Codebase, ms *labels.MatchSet) (m labels.Match, err error) {
		m.Label = labels.DepsNode
		packagePath := findPackageJSON(c)
		m.Valid = packagePath != ""
		if !m.Valid {
			return m, err
		}
		err = readPackageJSON(c, packagePath, &m)
		return m, err
	},
	func(c codebase.Codebase, ms *labels.MatchSet) (m labels.Match, err error) {
		m.Label = labels.TestJest
		m.Valid = hasDependency(ms, "jest")
		return m, err
	},
}

// npmPackageJSON for unmarshalling npm package.json files
type npmPackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

func findPackageJSON(c codebase.Codebase) string {
	file, _ := c.FindFile("package.json")
	if file != "" {
		return file
	}

	file, _ = c.FindFile("*/package.json")
	return file
}

func readPackageJSON(c codebase.Codebase, filePath string, match *labels.Match) error {
	var packageJSON npmPackageJSON

	fileContents, err := c.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileContents, &packageJSON)
	if err != nil {
		return err
	}

	match.MatchData.BasePath = path.Dir(filePath)
	match.Tasks = packageJSON.Scripts
	match.Dependencies = make(map[string]string)

	for k, v := range packageJSON.Dependencies {
		match.Dependencies[k] = v
	}
	for k, v := range packageJSON.DevDependencies {
		match.Dependencies[k] = v
	}

	return err
}

func hasDependency(ms *labels.MatchSet, dep string) bool {
	return (*ms)[labels.DepsNode].Dependencies[dep] != ""
}

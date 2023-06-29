package internal

import (
	"encoding/json"
	"errors"
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var PhpRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.DepsPhp
		label.Dependencies = make(map[string]string)

		composerPath, err := c.FindFile("composer.json")
		if err != nil && !errors.Is(err, codebase.NotFoundError) {
			return label, err

		}
		if composerPath != "" {
			label.Valid = true
			label.BasePath = path.Dir(composerPath)
		}

		err = readComposerFile(c, composerPath, &label)
		return label, err
	},
}

type composerJSON struct {
	Dependencies    map[string]string `json:"require"`
	DevDependencies map[string]string `json:"require-dev"`
}

func readComposerFile(c codebase.Codebase, filePath string, label *labels.Label) error {
	fileContents, err := c.ReadFile(filePath)
	if err != nil {
		return err
	}
	var composerDeps composerJSON
	err = json.Unmarshal(fileContents, &composerDeps)
	if err != nil {
		return err
	}
	label.Dependencies = make(map[string]string)
	for k, v := range composerDeps.Dependencies {
		label.Dependencies[k] = v
	}
	for k, v := range composerDeps.DevDependencies {
		label.Dependencies[k] = v
	}
	return nil
}

package internal

import (
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var EmptyRepoRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.EmptyRepo
		isEmpty, err := isRepoEmpty(c)

		if err == codebase.NotFoundError || isEmpty {
			label.Valid = true
			return label, nil
		}
		return label, err
	},
}

func isRepoEmpty(c codebase.Codebase) (bool, error) {
	files, err := c.ListFiles()
	if err != nil {
		return false, err
	}

	if len(files) == 0 {
		return true, nil
	}

	if len(files) == 1 {
		readme := strings.ToLower(strings.TrimSpace(files[0]))
		return readme == "readme.md", nil
	}

	return false, nil
}

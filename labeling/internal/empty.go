package internal

import (
	"errors"
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var EmptyRepoRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.EmptyRepo
		fileCount := 0
		_, err = c.FindFileMatching(func(path string) bool {
			if path != "." && strings.TrimSpace(strings.ToLower(path)) != "readme.md" {
				fileCount += 1
			}

			return false
		}, "*")

		if errors.Is(err, codebase.NotFoundError) && fileCount == 0 {
			label.Valid = true
			return label, nil
		}
		return label, err
	},
}

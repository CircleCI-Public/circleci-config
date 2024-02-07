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
		_, err = c.FindFileMatching(func(path string) bool {
			if path == "." || strings.TrimSpace(strings.ToLower(path)) == "readme.md" {
				return false
			}

			return true
		}, "*")

		if errors.Is(err, codebase.NotFoundError) {
			label.Valid = true
			return label, nil
		}
		return label, err
	},
}

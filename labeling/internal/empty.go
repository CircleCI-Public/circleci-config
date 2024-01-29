package internal

import (
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var EmptyRepoRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.EmptyRepo
		_, err = c.FindFileMatching(func(path string) bool {
			if path != "." {
				return true
			}
			return false
		}, "*")

		if err == codebase.NotFoundError {
			label.Valid = true
			return label, nil
		}
		return label, err
	},
}

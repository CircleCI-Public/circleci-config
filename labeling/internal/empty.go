package internal

import (
	"fmt"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var EmptyRepoRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.EmptyRepo
		result, err := c.FindFileMatching(func(path string) bool {
			if path == "." {
				return false
			}
			if path == "README.md" {
				return false
			}
			return true
		}, "*.*")
		fmt.Println("====> result", result, err)
		if err == codebase.NotFoundError {
			label.Valid = true
			fmt.Println("found empty repo")
			return label, nil
		}
		return label, err
	},
}

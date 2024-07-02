package internal

import (
	"errors"
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var ignoreList = []string{
	".",
	"readme.md",
	".git",
}

var EmptyRepoRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.EmptyRepo
		_, err = c.FindFileMatching(func(path string) bool {
			path = strings.TrimSpace(strings.ToLower(path))
			return !shouldIgnorePath(path)
		}, "*")

		if errors.Is(err, codebase.NotFoundError) {
			label.Valid = true
			return label, nil
		}
		return label, err
	},
}

func shouldIgnorePath(path string) bool {
	for _, token := range ignoreList {
		if path == token {
			return true
		}
	}

	return strings.HasPrefix(path, ".git/")
}

package internal

import "github.com/CircleCI-Public/circleci-config/labeling/codebase"

func hasPath(c codebase.Codebase, path string) bool {
	foundPath, _ := c.FindFile(path)
	return foundPath != ""
}

package internal

import (
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"go/parser"
	"go/token"
	"path"
)

var GoRules = []labels.Rule{
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.DepsGo
		goModPath, err := c.FindFile("go.mod")
		label.Valid = goModPath != ""
		label.BasePath = path.Dir(goModPath)
		return label, err
	},
	func(c codebase.Codebase, ls *labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.ArtifactGoExecutable
		if !(*ls)[labels.DepsGo].Valid {
			return label, err
		}

		label.Valid = containsMainGoFiles(c)
		return label, err
	},
}

func containsMainGoFiles(c codebase.Codebase) bool {
	_, err := c.FindFileMatching(
		func(path string) bool {
			return isMainPackageGoFile(c, path)
		},
		"*.go",
	)
	return err == nil
}

func isMainPackageGoFile(c codebase.Codebase, path string) bool {
	contents, innerErr := c.ReadFile(path)
	if innerErr != nil {
		return false
	}

	fileSet := token.NewFileSet()
	fileAst, err := parser.ParseFile(fileSet, "", contents, parser.PackageClauseOnly)
	return err == nil && fileAst.Name.Name == "main"
}

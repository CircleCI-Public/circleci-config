package labels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
)

const (
	ArtifactGoExecutable  = "artifact:go-executable"
	ArtifactRustCrate     = "artifact:rust-crate"
	DepsGo                = "deps:go"
	DepsJava              = "deps:java"
	DepsNode              = "deps:node"
	DepsPython            = "deps:python"
	DepsRuby              = "deps:ruby"
	DepsRust              = "deps:rust"
	PackageManagerPipenv  = "package_manager:pipenv"
	PackageManagerPoetry  = "package_manager:poetry"
	PackageManagerYarn    = "package_manager:yarn"
	PackageManagerGemspec = "package_manager:gemspec"
	TestJest              = "test:jest"
	ToolGradle            = "tool:gradle"
	FileManagePy          = "file:manage.py"
)

type LabelData struct {
	BasePath     string
	Dependencies map[string]string
	Tasks        map[string]string
	HasLockFile  bool
	Version      string
}

// Label is the result of applying a Rule
type Label struct {
	Key       string // string identifying the label, like "deps:go"
	Valid     bool   // If the rule applies, Valid = true
	LabelData        // LabelData rule-specific data for each label
}

func (label Label) String() string {
	if label.Valid {
		return fmt.Sprintf("%s:%s", label.Key, label.BasePath)
	} else {
		return fmt.Sprintf("!%s", label.Key)
	}
}

type LabelSet map[string]Label

func (ls LabelSet) String() string {
	labelsAsStrings := make([]string, len(ls))
	i := 0
	for _, v := range ls {
		labelsAsStrings[i] = v.String()
		i++
	}
	sort.Strings(labelsAsStrings)
	return strings.Join(labelsAsStrings, ",")
}

type Rule func(codebase.Codebase, LabelSet) (Label, error)

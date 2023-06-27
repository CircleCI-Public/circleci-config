package internal

import (
	"path"

	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

var RustRules = []labels.Rule{
	func(c codebase.Codebase, ls labels.LabelSet) (label labels.Label, err error) {
		label.Key = labels.DepsRust
		cargoTomlFile, err := c.FindFile("Cargo.toml", "cargo.toml")
		label.Valid = cargoTomlFile != ""
		label.BasePath = path.Dir(cargoTomlFile)
		return label, err
	},
}

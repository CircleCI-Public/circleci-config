package labeling

import (
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/internal"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

// ApplyRules applies the rules to a codebase.Codebase and returns a map of label key to
// valid codebase.Label
// Order of rules is relevant, higher "salience" rules should come first, i.e. later rules can
// depend on the LabelData of previous rules.
func ApplyRules(c codebase.Codebase, rules []labels.Rule) labels.LabelSet {
	ls := make(labels.LabelSet)
	for _, r := range rules {
		label, err := r(c, &ls)
		if err != nil {
			continue
		}

		if label.Valid {
			ls[label.Key] = label
		}
	}

	return ls
}

func ApplyAllRules(c codebase.Codebase) labels.LabelSet {
	allStacks := [][]labels.Rule{
		internal.NodeRules,
		internal.GoRules,
		internal.PythonRules,
		internal.RubyRules,
		// Add other stacks here
	}

	var allRules []labels.Rule
	for _, stack := range allStacks {
		allRules = append(allRules, stack...)
	}
	return ApplyRules(c, allRules)
}

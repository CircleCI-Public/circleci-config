package labeling

import (
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/internal"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

// ApplyRules applies the rules to a codebase.Codebase and returns a map of rule label to
// valid codebase.Match
// Order of rules is relevant, higher "salience" rules should come first, i.e. later rules can
// depend on the MatchData of previous rules.
func ApplyRules(c codebase.Codebase, rules []labels.Rule) labels.MatchSet {
	matches := make(labels.MatchSet)
	for _, r := range rules {
		match, err := r(c, &matches)
		if err != nil {
			continue
		}

		if match.Valid {
			matches[match.Label] = match
		}
	}

	return matches
}

func ApplyAllRules(c codebase.Codebase) labels.MatchSet {
	allStacks := [][]labels.Rule{
		internal.NodeRules,
		internal.GoRules,
		// Add other stacks here
	}

	var allRules []labels.Rule
	for _, stack := range allStacks {
		allRules = append(allRules, stack...)
	}
	return ApplyRules(c, allRules)
}

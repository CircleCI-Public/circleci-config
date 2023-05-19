package labels

import (
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
)

const (
	DepsNode = "deps:node"
	TestJest = "test:jest"
	DepsGo   = "deps:go"
)

type MatchData struct {
	BasePath     string
	Dependencies map[string]string
	Tasks        map[string]string
}

// Match is the result of applying a Rule
type Match struct {
	Label     string // Label of the matching rule
	Valid     bool   // If the rule matched, Valid = true
	MatchData        // MatchData rule-specific data for each match
}

type MatchSet map[string]Match

type Rule func(codebase.Codebase, *MatchSet) (Match, error)

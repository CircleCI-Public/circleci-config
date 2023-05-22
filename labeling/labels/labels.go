package labels

import (
	"fmt"
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"sort"
	"strings"
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

func (m Match) String() string {
	if m.Valid {
		return fmt.Sprintf("%s:%s", m.Label, m.BasePath)
	} else {
		return fmt.Sprintf("!%s", m.Label)
	}
}

type MatchSet map[string]Match

func (m MatchSet) String() string {
	matchStrings := make([]string, len(m))
	i := 0
	for _, v := range m {
		matchStrings[i] = v.String()
		i++
	}
	sort.Strings(matchStrings)
	return strings.Join(matchStrings, ",")
}

type Rule func(codebase.Codebase, *MatchSet) (Match, error)

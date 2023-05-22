package labels

import "testing"

func TestMatch_String(t *testing.T) {
	tests := []struct {
		name      string
		Label     string
		Valid     bool
		MatchData MatchData
		expected  string
	}{
		{
			name:      "invalid",
			Label:     "label",
			Valid:     false,
			MatchData: MatchData{},
			expected:  "!label",
		}, {
			name:  "valid",
			Label: "label",
			Valid: true,
			MatchData: MatchData{
				BasePath: ".",
			},
			expected: "label:.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Match{
				Label:     tt.Label,
				Valid:     tt.Valid,
				MatchData: tt.MatchData,
			}
			got := m.String()
			if got != tt.expected {
				t.Errorf("String() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestMatchSet_String(t *testing.T) {
	tests := []struct {
		name     string
		matchSet MatchSet
		expected string
	}{
		{
			name:     "empty",
			matchSet: MatchSet{},
			expected: "",
		}, {
			name: "3 matches, sorted",
			matchSet: MatchSet{
				DepsGo:   {Label: DepsGo, Valid: true},
				TestJest: {Label: TestJest, Valid: true},
				DepsNode: {Label: DepsNode, Valid: true}},
			expected: "deps:go:,deps:node:,test:jest:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.matchSet.String()
			if got != tt.expected {
				t.Errorf("String() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

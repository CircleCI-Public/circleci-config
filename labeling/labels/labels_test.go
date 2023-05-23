package labels

import "testing"

func TestLabel_String(t *testing.T) {
	tests := []struct {
		name      string
		Label     string
		valid     bool
		labelData LabelData
		expected  string
	}{
		{
			name:      "invalid",
			Label:     "label",
			valid:     false,
			labelData: LabelData{},
			expected:  "!label",
		}, {
			name:  "valid",
			Label: "label",
			valid: true,
			labelData: LabelData{
				BasePath: ".",
			},
			expected: "label:.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := Label{
				Key:       tt.Label,
				Valid:     tt.valid,
				LabelData: tt.labelData,
			}
			got := label.String()
			if got != tt.expected {
				t.Errorf("String() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestLabelSet_String(t *testing.T) {
	tests := []struct {
		name     string
		labels   LabelSet
		expected string
	}{
		{
			name:     "empty",
			labels:   LabelSet{},
			expected: "",
		}, {
			name: "3 labels, sorted",
			labels: LabelSet{
				DepsGo:   {Key: DepsGo, Valid: true},
				TestJest: {Key: TestJest, Valid: true},
				DepsNode: {Key: DepsNode, Valid: true}},
			expected: "deps:go:,deps:node:,test:jest:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.labels.String()
			if got != tt.expected {
				t.Errorf("String() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

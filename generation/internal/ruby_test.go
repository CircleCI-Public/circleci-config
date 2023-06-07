package internal

import (
	"reflect"
	"testing"

	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func Test_rubyImageVersion(t *testing.T) {
	tests := []struct {
		name            string
		labels          labels.LabelSet
		expectedVersion string
	}{
		{
			name: "version in gemfile",
			labels: labels.LabelSet{
				labels.DepsRuby: labels.Label{
					Key: labels.DepsRuby,
					LabelData: labels.LabelData{
						Dependencies: map[string]string{
							"ruby": "2.9.2",
						},
					},
				},
			},
			expectedVersion: "cimg/ruby:2.9.2-node",
		},
		{
			name: "no version in gemfile - use fallback",
			labels: labels.LabelSet{
				labels.DepsRuby: labels.Label{
					Key: labels.DepsRuby,
				},
			},
			expectedVersion: "cimg/ruby:3.2-node",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rubyImageVersion(tt.labels)
			if !reflect.DeepEqual(got, tt.expectedVersion) {
				t.Errorf("\n"+
					"got      %v\n"+
					"expected %v",
					got,
					tt.expectedVersion)
			}
		})
	}
}

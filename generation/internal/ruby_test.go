package internal

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/CircleCI-Public/circleci-config/config"
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

func Test_rubyInitialSteps(t *testing.T) {
	tests := []struct {
		name string
		ls   labels.LabelSet
		want []config.Step
	}{
		{
			name: "gemfile w/lockfile",
			ls: labels.LabelSet{
				labels.DepsRuby: labels.Label{
					Key: labels.DepsRuby,
					LabelData: labels.LabelData{
						BasePath:    ".",
						HasLockFile: true,
					},
				},
			},
			want: []config.Step{
				{Type: config.Checkout},
				{Type: config.OrbCommand, Command: "ruby/install-deps"},
			},
		},
		{
			name: "gemfile w/o lockfile",
			ls: labels.LabelSet{
				labels.DepsRuby: labels.Label{
					Key:   labels.DepsRuby,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath:    ".",
						HasLockFile: false,
					},
				},
				labels.PackageManagerGemspec: labels.Label{
					Key:   labels.PackageManagerGemspec,
					Valid: true,
					LabelData: labels.LabelData{
						Path: "./my.gemspec",
					},
				},
			},
			want: []config.Step{
				{Type: config.Checkout},
				{Type: config.Run,
					Command: "bundle install",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rubyInitialSteps(tt.ls); !reflect.DeepEqual(got, tt.want) {
				fmt.Println("want", tt.want)
				fmt.Println("got", got)
				t.Errorf("rubyInitialSteps() = %v, want %v", got, tt.want)
			}
		})
	}
}

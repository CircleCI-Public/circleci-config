package internal

import (
	"reflect"
	"testing"

	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"github.com/google/go-cmp/cmp"
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
				t.Errorf("rubyInitialSteps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateRubyJobs(t *testing.T) {
	type args struct {
		ls labels.LabelSet
	}
	tests := []struct {
		name     string
		args     args
		wantJobs []*Job
	}{
		{
			name: "gemfile has rake",
			args: args{ls: labels.LabelSet{
				labels.DepsRuby: labels.Label{
					Valid: true,
					LabelData: labels.LabelData{
						Dependencies: map[string]string{"rake": "true"},
						HasLockFile:  true,
					},
				}}},
			wantJobs: []*Job{
				{
					Job: config.Job{
						Name:             "test-ruby",
						Comment:          "Install gems, run rake tests",
						DockerImages:     []string{"cimg/ruby:3.2-node"},
						WorkingDirectory: "~/project",
						Steps: []config.Step{
							{
								Path: "~/project",
								Type: config.Checkout,
							},
							{
								Command: "ruby/install-deps",
								Type:    config.OrbCommand,
							},
							{
								Type:    config.Run,
								Name:    "rake test",
								Command: "bundle exec rake test",
							},
						}},
					Type: TestJob,
					Orbs: map[string]string{"ruby": "circleci/ruby@2.0.1"},
				},
			},
		},
		{
			name: "gemfile has rails w/o rake or rspec",
			args: args{ls: labels.LabelSet{
				labels.DepsRuby: labels.Label{
					Valid: true,
					LabelData: labels.LabelData{
						Dependencies: map[string]string{"rails": "true"},
						HasLockFile:  true,
					},
				}}},
			wantJobs: []*Job{
				{
					Job: config.Job{
						Name:             "test-ruby",
						Comment:          "Install gems, run rails tests",
						DockerImages:     []string{"cimg/ruby:3.2-node"},
						WorkingDirectory: "~/project",
						Steps: []config.Step{
							{
								Path: "~/project",
								Type: config.Checkout,
							},
							{
								Command: "ruby/install-deps",
								Type:    config.OrbCommand,
							},
							{
								Type:    config.Run,
								Name:    "rails test",
								Command: "bundle exec rails test",
							},
						}},
					Type: TestJob,
					Orbs: map[string]string{"ruby": "circleci/ruby@2.0.1"},
				},
			},
		},
		{
			name: "gemfile has rspec",
			args: args{ls: labels.LabelSet{
				labels.DepsRuby: labels.Label{
					Valid: true,
					LabelData: labels.LabelData{
						Dependencies: map[string]string{"rspec": "true"},
						HasLockFile:  true,
					},
				}}},
			wantJobs: []*Job{
				{
					Job: config.Job{
						Name:             "test-ruby",
						Comment:          "Install gems, run rspec tests",
						DockerImages:     []string{"cimg/ruby:3.2-node"},
						WorkingDirectory: "~/project",
						Environment:      map[string]string{"RAILS_ENV": "test"},
						Steps: []config.Step{
							{
								Path: "~/project",
								Type: config.Checkout,
							},
							{
								Command: "ruby/install-deps",
								Type:    config.OrbCommand,
							},
							{
								Type:    config.Run,
								Name:    "rspec test",
								Command: "bundle exec rspec",
							},
						}},
					Type: TestJob,
					Orbs: map[string]string{"ruby": "circleci/ruby@2.0.1"},
				},
			},
		},
		{
			name: "gemspec has rake",
			args: args{ls: labels.LabelSet{
				labels.DepsRuby: labels.Label{
					Valid: true,
				},
				labels.PackageManagerGemspec: labels.Label{
					Valid: true,
					LabelData: labels.LabelData{
						Dependencies: map[string]string{"rspec": "true"},
					},
				}}},
			wantJobs: []*Job{
				{
					Job: config.Job{
						Name:             "test-ruby",
						Comment:          "Install gems, run rspec tests",
						DockerImages:     []string{"cimg/ruby:3.2-node"},
						WorkingDirectory: "~/project",
						Environment:      map[string]string{"RAILS_ENV": "test"},
						Steps: []config.Step{
							{
								Path: "~/project",
								Type: config.Checkout,
							},
							{
								Command: "bundle install",
								Type:    config.Run,
							},
							{
								Type:    config.Run,
								Name:    "rspec test",
								Command: "bundle exec rspec",
							},
						}},
					Type: TestJob,
					Orbs: map[string]string{"ruby": "circleci/ruby@2.0.1"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJobs := GenerateRubyJobs(tt.args.ls)
			diff := cmp.Diff(tt.wantJobs, gotJobs)
			if diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

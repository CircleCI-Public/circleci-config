package internal

import (
	"testing"

	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"github.com/google/go-cmp/cmp"
)

func TestGeneratePHPJobs(t *testing.T) {
	type args struct {
		ls labels.LabelSet
	}
	tests := []struct {
		name     string
		args     args
		wantJobs []*Job
	}{
		{
			name: "composer file has phpunit",
			args: args{ls: labels.LabelSet{
				labels.DepsPhp: labels.Label{
					Valid: true,
					LabelData: labels.LabelData{
						Dependencies: map[string]string{"phpunit/phpunit": "~4.2"},
						HasLockFile:  true,
					},
				}}},
			wantJobs: []*Job{
				{
					Job: config.Job{
						Name:             "test-php",
						Comment:          "Install php packages and run tests",
						WorkingDirectory: "",
						DockerImages:     []string{"cimg/php:8.2.7-node"},
						Steps: []config.Step{
							{
								Path: "~/project",
								Type: config.Checkout,
							},
							{
								Command: "php/install-packages",
								Type:    config.OrbCommand,
							},
							{
								Type:    config.Run,
								Name:    "run tests",
								Command: "./vendor/bin/phpunit",
							},
						}},
					Type: TestJob,
					Orbs: map[string]string{"php": "circleci/php@1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJobs := GeneratePHPJobs(tt.args.ls)
			diff := cmp.Diff(tt.wantJobs, gotJobs)
			if diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

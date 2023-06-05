package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func GenerateRubyJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.PackageManagerBundler].Valid {
		return nil
	}

	if ls[labels.TestRSpec].Valid {
		jobs = append(jobs, rspecJob(ls))
	}

	return jobs
}

func rubyInitialSteps(ls labels.LabelSet) []config.Step {
	steps := initialSteps(ls[labels.PackageManagerBundler])

	steps = append(steps, config.Step{
		Type:    config.OrbCommand,
		Command: "ruby/install-deps",
	})
	return steps
}

const rubyOrb = "circleci/ruby@1.1.0"
const rubyImage = "cimg/ruby:2.7-node"
const postgresImage = "circleci/postgres:9.5-alpine"

func rspecJob(ls labels.LabelSet) *Job {
	steps := rubyInitialSteps(ls)
	steps = append(steps,
		config.Step{
			Type:    config.Run,
			Name:    "wait for DB",
			Command: "dockerize -wait tcp://localhost:5432 -timeout 1m"},
		config.Step{
			Type:    config.Run,
			Name:    "Database setup",
			Command: "bundle exec rake db:test:prepare"},
		config.Step{

			Type:    config.OrbCommand,
			Name:    "rspec test",
			Command: "ruby/rspec-test"})

	return &Job{
		Job: config.Job{
			Name:         "test-ruby",
			Comment:      "Install gems, run rspec tests",
			Steps:        steps,
			DockerImages: []string{rubyImage, postgresImage},
			Environment: map[string]string{
				"RAILS_ENV": "test"},
		},
		Orbs: map[string]string{
			"ruby": rubyOrb,
		},
	}
}
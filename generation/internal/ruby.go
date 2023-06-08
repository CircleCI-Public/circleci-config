package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func GenerateRubyJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.DepsRuby].Valid {
		return nil
	}

	if ls[labels.DepsRuby].Dependencies["rspec"] == "true" {
		jobs = append(jobs, rspecJob(ls))
	}

	return jobs
}

func rubyInitialSteps(ls labels.LabelSet) []config.Step {
	steps := initialSteps(ls[labels.DepsRuby])

	steps = append(steps, config.Step{
		Type:    config.OrbCommand,
		Command: "ruby/install-deps",
	})
	return steps
}

const rubyOrb = "circleci/ruby@1.1.0"
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

	images := []string{rubyImageVersion(ls)}
	if ls[labels.DepsRuby].Dependencies["pg"] == "true" {
		images = append(images, postgresImage)
	}

	return &Job{
		Job: config.Job{
			Name:         "test-ruby",
			Comment:      "Install gems, run rspec tests",
			Steps:        steps,
			DockerImages: images,
			Environment: map[string]string{
				"RAILS_ENV": "test"},
		},
		Orbs: map[string]string{
			"ruby": rubyOrb,
		},
	}
}

const rubyFallbackVersion = "3.2"

// Construct the ruby image tag based on the ruby version
func rubyImageVersion(ls labels.LabelSet) string {
	prefix := "cimg/ruby:"
	suffix := "-node"
	version := rubyFallbackVersion

	gemfileVersion := ls[labels.DepsRuby].Dependencies["ruby"]
	if gemfileVersion != "" {
		version = gemfileVersion
	}

	return prefix + version + suffix
}

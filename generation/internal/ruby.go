package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func GenerateRubyJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.DepsRuby].Valid {
		return nil
	}

	if hasGem(ls, "rake") {
		jobs = append(jobs, rakeJob(ls))
	} else if hasGem(ls, "rspec") {
		jobs = append(jobs, rspecJob(ls))
	}
	return jobs
}

func hasGem(ls labels.LabelSet, gem string) bool {
	for _, label := range []string{labels.DepsRuby, labels.PackageManagerGemspec} {
		if ls[label].Valid == true && ls[label].LabelData.Dependencies[gem] != "" {
			return true
		}
	}
	return false
}

func rubyInitialSteps(ls labels.LabelSet) []config.Step {

	checkout := checkoutStep(ls[labels.DepsRuby])

	installDeps := config.Step{Type: config.OrbCommand, Command: "ruby/install-deps"}

	// ruby orb requires Gemfile.lockfile, so revert to basic bundle command when not found
	if !ls[labels.DepsRuby].LabelData.HasLockFile && ls[labels.PackageManagerGemspec].Valid == true {
		installDeps = config.Step{Type: config.Run, Command: "bundle install"}
	}
	return []config.Step{checkout, installDeps}
}

const rubyOrb = "circleci/ruby@2.0.1"
const postgresImage = "circleci/postgres:9.5-alpine"

func rspecJob(ls labels.LabelSet) *Job {
	steps := rubyInitialSteps(ls)
	images := []string{rubyImageVersion(ls)}

	if ls[labels.DepsRuby].Dependencies["pg"] == "true" {
		images = append(images, postgresImage)

		steps = append(steps,
			config.Step{
				Type:    config.Run,
				Name:    "wait for DB",
				Command: "dockerize -wait tcp://localhost:5432 -timeout 1m"},
			config.Step{
				Type:    config.Run,
				Name:    "Database setup",
				Command: "bundle exec rake db:test:prepare"})
	}

	if hasGem(ls, "rspec_junit_formatter") {
		steps = append(steps,
			config.Step{
				Type:    config.OrbCommand,
				Name:    "rspec test",
				Command: "ruby/rspec-test"})
	} else {
		steps = append(steps,
			config.Step{
				Type:    config.Run,
				Name:    "rspec test",
				Command: "bundle exec rspec"})
	}

	return &Job{
		Job: config.Job{
			Name:             "test-ruby",
			Comment:          "Install gems, run rspec tests",
			Steps:            steps,
			DockerImages:     images,
			WorkingDirectory: workingDirectory(ls[labels.DepsRuby]),
			Environment: map[string]string{
				"RAILS_ENV": "test"},
		},
		Orbs: map[string]string{
			"ruby": rubyOrb,
		},
	}
}

func rakeJob(ls labels.LabelSet) *Job {
	steps := rubyInitialSteps(ls)
	steps = append(steps,
		config.Step{
			Type:    config.Run,
			Name:    "rake test",
			Command: "bundle exec rake test",
		})

	return &Job{
		Job: config.Job{
			Name:             "test-ruby",
			Comment:          "Install gems, run rake tests",
			Steps:            steps,
			DockerImages:     []string{rubyImageVersion(ls)},
			WorkingDirectory: workingDirectory(ls[labels.DepsRuby]),
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

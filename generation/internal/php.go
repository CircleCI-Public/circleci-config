package internal

import (
	"strings"

	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func GeneratePHPJobs(ls labels.LabelSet) []*Job {
	if !ls[labels.DepsPhp].Valid {
		return nil
	}

	jobs := []*Job{}

	if hasPhpLib(ls, "phpunit/phpunit") {
		jobs = append(jobs, phpTestJob(ls))
	}
	return jobs
}

func hasPhpLib(ls labels.LabelSet, lib string) bool {
	if ls[labels.DepsPhp].Dependencies[lib] != "" {
		return true
	}
	return false
}

func initialPhpSteps(ls labels.LabelSet) []config.Step {
	checkout := checkoutStep(ls[labels.DepsPhp])
	installPackages := config.Step{
		Type:    config.OrbCommand,
		Command: "php/install-packages",
	}
	return []config.Step{checkout, installPackages}
}

func phpImageVersion(composerVersion string) string {
	version := "8.2.7"
	if composerVersion != "" {
		version = strings.TrimPrefix(composerVersion, "~")
		version = strings.TrimPrefix(version, "^")
		version = strings.TrimSuffix(version, ".*")
	}
	return "cimg/php:" + version + "-node"
}

func phpTestJob(ls labels.LabelSet) *Job {
	steps := initialPhpSteps(ls)
	steps = append(steps, config.Step{
		Type:    config.Run,
		Name:    "run tests",
		Command: "./vendor/bin/phpunit",
	})
	return &Job{
		Job: config.Job{
			Name:         "test-php",
			Comment:      "Install php packages and run tests",
			Steps:        steps,
			DockerImages: []string{phpImageVersion(ls[labels.DepsPhp].Dependencies["php"])},
		},
		Orbs: map[string]string{
			"php": "circleci/php@1",
		},
	}
}

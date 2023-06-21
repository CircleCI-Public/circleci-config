package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const rustOrb = "circleci/rust@1.6.0"

func rustInitialSteps(ls labels.LabelSet) []config.Step {
	return []config.Step{checkoutStep(ls[labels.DepsRust])}
}

func rustTestJob(ls labels.LabelSet) *Job {
	steps := rustInitialSteps(ls)

	steps = append(steps, config.Step{
		Type:    config.OrbCommand,
		Command: "rust/test",
	})

	return &Job{
		Job: config.Job{
			Name:             "test-rust",
			Comment:          "Run tests using the rust orb",
			Executor:         "rust/default",
			WorkingDirectory: workingDirectory(ls[labels.DepsRust]),
			Steps:            steps,
		},
		Orbs: map[string]string{"rust": rustOrb},
		Type: TestJob,
	}
}

func GenerateRustJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.DepsRust].Valid {
		return nil
	}

	return append(jobs, rustTestJob(ls))
}

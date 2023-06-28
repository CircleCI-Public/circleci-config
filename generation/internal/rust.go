package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const rustDockerImage = "cimg/rust:1.70"
const cargoCacheKey = `cargo-{{ checksum "Cargo.lock" }}`

func rustInitialSteps(ls labels.LabelSet) []config.Step {
	return []config.Step{checkoutStep(ls[labels.DepsRust]), {
		Type:     config.RestoreCache,
		CacheKey: cargoCacheKey,
	}}
}

func rustTestJob(ls labels.LabelSet) *Job {
	steps := rustInitialSteps(ls)

	steps = append(steps, []config.Step{
		{
			Type:    config.Run,
			Command: "cargo test",
		},
		{
			Type:     config.SaveCache,
			CacheKey: cargoCacheKey,
			Path:     "~/.cargo",
		},
	}...)

	return &Job{
		Job: config.Job{
			Name:             "test-rust",
			DockerImages:     []string{rustDockerImage},
			WorkingDirectory: workingDirectory(ls[labels.DepsRust]),
			Steps:            steps,
		},
		Type: TestJob,
	}
}

func GenerateRustJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.DepsRust].Valid {
		return nil
	}

	return append(jobs, rustTestJob(ls))
}

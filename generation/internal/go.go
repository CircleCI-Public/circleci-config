package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const goCacheKey = `go-mod-{{ checksum "go.sum" }}`

func goTestJob(matches labels.MatchSet) *Job {
	steps := initialSteps(matches[labels.DepsGo])

	steps = append(steps, []config.Step{
		{
			Type:     config.RestoreCache,
			CacheKey: goCacheKey,
		}, {
			Type:    config.Run,
			Name:    "Download Go modules",
			Command: "go mod download",
		}, {
			Type:     config.SaveCache,
			CacheKey: goCacheKey,
			Path:     "/home/circleci/go/pkg/mod",
		}, {
			Type:    config.Run,
			Name:    "Run go vet",
			Command: "go vet ./...",
		}, {
			Type:    config.Run,
			Name:    "Run tests",
			Command: "gotestsum --junitfile junit.xml",
		}, {
			Type: config.StoreTestResults,
			Path: "junit.xml",
		}}...)

	return &Job{
		Job: config.Job{
			Name:        "test-go",
			Comment:     "Install go modules, run go vet and tests",
			DockerImage: "cimg/go",
			Steps:       steps,
		},
		Type: TestJob,
	}
}

func GenerateGoJobs(matches labels.MatchSet) []*Job {
	if !matches[labels.DepsGo].Valid {
		return nil
	}
	return []*Job{goTestJob(matches)}
}

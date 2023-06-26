package internal

import (
	"fmt"

	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func goInitialSteps(ls labels.LabelSet) []config.Step {
	const goCacheKey = `go-mod-{{ checksum "go.sum" }}`

	return []config.Step{
		checkoutStep(ls[labels.DepsGo]),
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
		}}
}

func goTestJob(ls labels.LabelSet) *Job {
	steps := goInitialSteps(ls)

	steps = append(steps, []config.Step{{
		Type:    config.Run,
		Name:    "Run tests",
		Command: "gotestsum --junitfile junit.xml",
	}, {
		Type: config.StoreTestResults,
		Path: "junit.xml",
	}}...)

	return &Job{
		Job: config.Job{
			Name:             "test-go",
			Comment:          "Install go modules and run tests",
			DockerImages:     []string{"cimg/go:1.20"},
			WorkingDirectory: workingDirectory(ls[labels.DepsGo]),
			Steps:            steps,
		},
		Type: TestJob,
	}
}

func goBuildJob(ls labels.LabelSet) *Job {
	steps := goInitialSteps(ls)

	steps = append(steps,
		createArtifactsDirStep,
		config.Step{
			Type:    config.Run,
			Name:    "Build executables",
			Command: fmt.Sprintf("go build -o %s ./...", artifactsPath),
		},
		storeArtifactsStep("executables"))

	return &Job{
		Job: config.Job{
			Name:         "build-go-executables",
			Comment:      "Build go executables and store them as artifacts",
			DockerImages: []string{"cimg/go:1.20"},
			Steps:        steps,
		},
		Type: ArtifactJob,
	}
}

func GenerateGoJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.DepsGo].Valid {
		return nil
	}

	jobs = append(jobs, goTestJob(ls))

	if ls[labels.ArtifactGoExecutable].Valid {
		jobs = append(jobs, goBuildJob(ls))
	}

	return jobs
}

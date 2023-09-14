package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const pythonOrb = "circleci/python@2"

func defaultSteps(l labels.Label, hasManagePy bool) []config.Step {
	steps := []config.Step{
		{
			Type:    config.OrbCommand,
			Command: "python/install-packages",
		},
	}

	if hasManagePy {
		steps = append(steps, config.Step{
			Name:    "Run tests",
			Type:    config.Run,
			Command: "python manage.py test",
		})
		return steps
	}

	steps = append(steps, []config.Step{
		{
			Name:    "Run tests",
			Type:    config.Run,
			Command: "pytest --junitxml=junit.xml",
		}}...)

	return steps
}

func pythonTestJob(ls labels.LabelSet) *Job {
	steps := []config.Step{checkoutStep(ls[labels.DepsPython])}
	hasManagePy := ls[labels.FileManagePy].Valid

	steps = append(steps, defaultSteps(ls[labels.DepsPython], hasManagePy)...)

	if !hasManagePy {
		// Store test results
		steps = append(steps, config.Step{
			Type: config.StoreTestResults,
			Path: "junit.xml",
		})
	}

	return &Job{
		Job: config.Job{
			Name:             "test-python",
			Comment:          "Install dependencies and run tests",
			DockerImages:     []string{pythonImageVersion(ls)},
			WorkingDirectory: workingDirectory(ls[labels.DepsPython]),
			Steps:            steps,
		},
		Type: TestJob,
		Orbs: map[string]string{
			"python": pythonOrb,
		},
	}
}

func GeneratePythonJobs(ls labels.LabelSet) []*Job {
	if !ls[labels.DepsPython].Valid {
		return nil
	}

	return []*Job{
		pythonTestJob(ls),
	}
}

const pythonFallbackVersion = "3.8"

// Construct the python image tag based on the python version
func pythonImageVersion(ls labels.LabelSet) string {
	prefix := "cimg/python:"
	suffix := "-node"
	version := pythonFallbackVersion

	pythonVersion := ls[labels.DepsPython].Dependencies["python"]
	if pythonVersion != "" {
		version = pythonVersion
	}

	return prefix + version + suffix
}

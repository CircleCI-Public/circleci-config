package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const pythonOrb = "circleci/python@2"

func pipSteps(l labels.Label) []config.Step {
	return []config.Step{
		{
			Type:    config.OrbCommand,
			Command: "python/install-packages",
			Parameters: config.OrbCommandParameters{
				"app-dir":     l.BasePath,
				"pkg-manager": "pip",
			},
		},
		{
			Type:    config.OrbCommand,
			Command: "python/install-packages",
			Parameters: config.OrbCommandParameters{
				"args":        "pytest",
				"pkg-manager": "pip",
				"pypi-cache":  "false",
			},
		},
		{
			Name:    "Run tests",
			Type:    config.Run,
			Command: "mkdir test-results && pytest  --junitxml=test-results/junit.xml",
		},
	}
}

func pipenvSteps(l labels.Label) []config.Step {
	return []config.Step{
		{
			Type:    config.OrbCommand,
			Command: "python/install-packages",
			Parameters: config.OrbCommandParameters{
				"app-dir":     l.BasePath,
				"pkg-manager": "pipenv",
			},
		},
		{
			Name:    "Run tests",
			Type:    config.Run,
			Command: "mkdir test-results && pipenv run pytest --junitxml=test-results/junit.xml",
		},
	}
}

func poetrySteps(l labels.Label) []config.Step {
	return []config.Step{
		{
			Type:    config.OrbCommand,
			Command: "python/install-packages",
			Parameters: config.OrbCommandParameters{
				"app-dir":     l.BasePath,
				"pkg-manager": "poetry",
			},
		},
		{
			Name:    "Run tests",
			Type:    config.Run,
			Command: "mkdir test-results && poetry run pytest --junitxml=test-results/junit.xml",
		},
	}
}

func pythonTestJob(ls labels.LabelSet) *Job {
	steps := initialSteps(ls[labels.DepsPython])

	// Support for different package managers
	switch {
	case ls[labels.PackageManagerPipenv].Valid:
		steps = append(steps, pipenvSteps(ls[labels.PackageManagerPipenv])...)
	case ls[labels.PackageManagerPoetry].Valid:
		steps = append(steps, poetrySteps(ls[labels.PackageManagerPoetry])...)
	default:
		steps = append(steps, pipSteps(ls[labels.DepsPython])...)
	}

	// Store test results
	steps = append(steps, config.Step{
		Type:    config.OrbCommand,
		Command: "store_test_results",
		Parameters: config.OrbCommandParameters{
			"path": "test-results",
		},
	})

	return &Job{
		Job: config.Job{
			Name:     "test-python",
			Comment:  "Install dependencies and run tests",
			Executor: "python/default",
			Steps:    steps,
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

package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const pythonOrb = "circleci/python@2"

func pipSteps(l labels.Label, hasManagePy bool) []config.Step {
	steps := []config.Step{
		{
			Type:    config.OrbCommand,
			Command: "python/install-packages",
			Parameters: config.OrbCommandParameters{
				"pkg-manager": "pip",
			},
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
			Command: "pytest --junitxml=junit.xml",
		}}...)

	return steps
}

func pipenvSteps(l labels.Label, hasManagePy bool) []config.Step {
	steps := []config.Step{
		{
			Type:    config.OrbCommand,
			Command: "python/install-packages",
			Parameters: config.OrbCommandParameters{
				"pkg-manager": "pipenv",
			},
		},
	}

	if hasManagePy {
		steps = append(steps, config.Step{
			Name:    "Run tests",
			Type:    config.Run,
			Command: "pipenv run python manage.py test",
		})
		return steps
	}

	steps = append(steps, config.Step{
		Name:    "Run tests",
		Type:    config.Run,
		Command: "pipenv run pytest --junitxml=junit.xml",
	})

	return steps
}

func poetrySteps(l labels.Label, hasManagerPy bool) []config.Step {
	steps := []config.Step{
		{
			Type:    config.OrbCommand,
			Command: "python/install-packages",
			Parameters: config.OrbCommandParameters{
				"pkg-manager": "poetry",
			},
		},
	}

	if hasManagerPy {
		steps = append(steps, config.Step{
			Name:    "Run tests",
			Type:    config.Run,
			Command: "poetry run python manage.py test",
		})
		return steps
	}

	steps = append(steps, config.Step{
		Name:    "Run tests",
		Type:    config.Run,
		Command: "poetry run pytest --junitxml=junit.xml",
	})

	return steps
}

func pythonTestJob(ls labels.LabelSet) *Job {
	steps := []config.Step{checkoutStep(ls[labels.DepsPython])}
	hasManagePy := ls[labels.FileManagePy].Valid

	// Support for different package managers
	switch {
	case ls[labels.PackageManagerPipenv].Valid:
		steps = append(steps, pipenvSteps(ls[labels.PackageManagerPipenv], hasManagePy)...)
	case ls[labels.PackageManagerPoetry].Valid:
		steps = append(steps, poetrySteps(ls[labels.PackageManagerPoetry], hasManagePy)...)
	default:
		steps = append(steps, pipSteps(ls[labels.DepsPython], hasManagePy)...)
	}

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
			Executor:         "python/default",
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

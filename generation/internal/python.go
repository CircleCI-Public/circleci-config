package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const pythonOrb = "circleci/python@2"

func testSteps(ls labels.LabelSet) []config.Step {
	hasManagePy := ls[labels.FileManagePy].Valid
	hasSetupPy := ls[labels.FileSetupPy].Valid
	hasTox := ls[labels.TestTox].Valid
	hasPipenv := ls[labels.PackageManagerPipenv].Valid
	hasPoetry := ls[labels.PackageManagerPoetry].Valid

	packageManager := "auto"
	installArgs := ""
	commandPrefix := ""

	if hasSetupPy {
		packageManager = "pip-dist"
	}

	if hasPipenv {
		packageManager = "pipenv"
		commandPrefix = "pipenv run "
		installArgs = "--dev"
	}

	if hasPoetry {
		packageManager = "poetry"
		commandPrefix = "poetry run "
	}

	installParams := config.OrbCommandParameters{}
	if packageManager != "auto" {
		installParams["pkg-manager"] = packageManager
	}
	if installArgs != "" {
		installParams["args"] = installArgs
	}

	steps := []config.Step{
		{
			Type:       config.OrbCommand,
			Command:    "python/install-packages",
			Parameters: installParams,
		},
	}

	if hasManagePy {
		steps = append(steps, config.Step{
			Name:    "Run tests",
			Type:    config.Run,
			Command: commandPrefix + "python manage.py test",
		})
		return steps
	}

	if hasTox {
		steps = append(steps,
			config.Step{
				Name:    "Install tox",
				Type:    config.OrbCommand,
				Command: "python/install-packages",
				Parameters: config.OrbCommandParameters{
					"args":        "tox",
					"pkg-manager": packageManager},
			},
			config.Step{
				Name:    "Run tests",
				Type:    config.Run,
				Command: commandPrefix + "tox",
			},
			config.Step{
				Type: config.StoreTestResults,
				Path: "junit.xml",
			},
		)
		return steps
	}

	// Run pytest via package manager (or directly)
	steps = append(steps, []config.Step{
		{
			Name:    "Run tests",
			Type:    config.Run,
			Command: commandPrefix + "pytest --junitxml=junit.xml || ((($? == 5)) && echo 'Did not find any tests to run.')",
		},
		{
			Type: config.StoreTestResults,
			Path: "junit.xml",
		}}...)

	return steps
}

func pythonTestJob(ls labels.LabelSet) *Job {
	steps := []config.Step{
		checkoutStep(ls[labels.DepsPython]),
	}
	steps = append(steps, testSteps(ls)...)

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

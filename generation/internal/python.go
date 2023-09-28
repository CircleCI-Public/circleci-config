package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const pythonOrb = "circleci/python@2"

type pythonPackageManager interface {
	installPackages() config.Step
	installPackage(pkg string) config.Step
	run(command string) string
}

type pythonTestRunner interface {
	testSteps(mgr pythonPackageManager) []config.Step
}

func pythonTestJob(ls labels.LabelSet) *Job {
	steps := []config.Step{
		checkoutStep(ls[labels.DepsPython]),
	}

	var mgr pythonPackageManager = defaultManager{}
	switch {
	case ls[labels.FileSetupPy].Valid:
		mgr = setuptools{}
	case ls[labels.PackageManagerPipenv].Valid:
		mgr = pipenv{}
	case ls[labels.PackageManagerPoetry].Valid:
		mgr = poetry{}
	}

	steps = append(steps, mgr.installPackages())

	var testRunner pythonTestRunner = pytest{}
	switch {
	case ls[labels.FileManagePy].Valid:
		testRunner = manage{}
	case ls[labels.TestTox].Valid:
		testRunner = tox{}
	}
	steps = append(steps, testRunner.testSteps(mgr)...)

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

func pythonBuildJob(ls labels.LabelSet) *Job {

	steps := []config.Step{
		checkoutStep(ls[labels.DepsPython]),
		createArtifactsDirStep,
		{
			Type:    config.OrbCommand,
			Name:    "build package",
			Command: "python/dist",
		},
		{
			Type:        config.StoreArtifacts,
			Path:        "dist",
			Destination: "~/artifacts",
		},
	}
	return &Job{
		Job: config.Job{
			Name:         "build-package",
			Comment:      "build python package",
			DockerImages: []string{pythonImageVersion(ls)},
			Steps:        steps,
		},
		Type: ArtifactJob,
		Orbs: map[string]string{
			"python": pythonOrb,
		},
	}
}

func GeneratePythonJobs(ls labels.LabelSet) []*Job {
	if !ls[labels.DepsPython].Valid {
		return nil
	}
	jobs := []*Job{
		pythonTestJob(ls),
	}
	if ls[labels.FileSetupPy].Valid {
		jobs = append(jobs, pythonBuildJob(ls))
	}
	return jobs
}

type defaultManager struct{}

func (d defaultManager) run(command string) string {
	return command
}

func (d defaultManager) installPackages() config.Step {
	return config.Step{
		Type:    config.OrbCommand,
		Command: "python/install-packages",
	}
}
func (d defaultManager) installPackage(pkg string) config.Step {
	return config.Step{
		Type:    config.OrbCommand,
		Command: "python/install-packages",
		Parameters: config.OrbCommandParameters{
			"args": pkg,
		},
	}
}

type setuptools struct{}

func (s setuptools) installPackages() config.Step {
	return config.Step{
		Type:    config.OrbCommand,
		Command: "python/install-packages",
		Parameters: config.OrbCommandParameters{
			"pkg-manager": "pip-dist",
		},
	}
}
func (s setuptools) installPackage(pkg string) config.Step {
	return config.Step{
		Type:    config.OrbCommand,
		Command: "python/install-packages",
		Parameters: config.OrbCommandParameters{
			"pkg-manager": "pip-dist",
			"args":        pkg,
		},
	}
}
func (s setuptools) run(command string) string {
	return command
}

type pipenv struct{}

func (p pipenv) installPackages() config.Step {
	return config.Step{
		Type:    config.OrbCommand,
		Command: "python/install-packages",
		Parameters: config.OrbCommandParameters{
			"args":        "--dev",
			"pkg-manager": "pipenv",
		},
	}
}

func (p pipenv) installPackage(pkg string) config.Step {
	return config.Step{
		Type:    config.OrbCommand,
		Command: "python/install-packages",
		Parameters: config.OrbCommandParameters{
			"args":        pkg,
			"pkg-manager": "pipenv",
		},
	}
}

func (p pipenv) run(command string) string {
	return "pipenv run " + command
}

type poetry struct{}

func (p poetry) installPackages() config.Step {
	return config.Step{
		Type:    config.OrbCommand,
		Command: "python/install-packages",
		Parameters: config.OrbCommandParameters{
			"pkg-manager": "poetry",
		},
	}
}

func (p poetry) installPackage(pkg string) config.Step {
	return config.Step{
		Type:    config.OrbCommand,
		Command: "python/install-packages",
		Parameters: config.OrbCommandParameters{
			"pkg-manager": "poetry",
			"args":        pkg,
		},
	}
}

func (p poetry) run(command string) string {
	return "poetry run " + command
}

type manage struct{}

func (m manage) testSteps(mgr pythonPackageManager) []config.Step {
	return []config.Step{
		{
			Name:    "Run tests",
			Type:    config.Run,
			Command: mgr.run("python manage.py test"),
		},
	}

}

type pytest struct{}

func (p pytest) testSteps(mgr pythonPackageManager) []config.Step {
	return []config.Step{
		{
			Name:    "Run tests",
			Type:    config.Run,
			Command: mgr.run("pytest --junitxml=junit.xml || ((($? == 5)) && echo 'Did not find any tests to run.')"),
		},
		{
			Type: config.StoreTestResults,
			Path: "junit.xml",
		},
	}
}

type tox struct{}

func (t tox) testSteps(mgr pythonPackageManager) []config.Step {
	return []config.Step{
		mgr.installPackage("tox"),
		{
			Name:    "Run tests",
			Type:    config.Run,
			Command: mgr.run("tox"),
		},
		{
			Type: config.StoreTestResults,
			Path: "junit.xml",
		},
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

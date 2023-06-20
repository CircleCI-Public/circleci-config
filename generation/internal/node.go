package internal

import (
	"fmt"

	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const nodeOrb = "circleci/node@5"

func npmTaskDefined(ls labels.LabelSet, task string) bool {
	return ls[labels.DepsNode].Tasks[task] != ""
}

func nodeTestSteps(ls labels.LabelSet, packageManager string) []config.Step {
	hasJestLabel := ls[labels.TestJest].Valid

	if npmTaskDefined(ls, "test:ci") {
		return []config.Step{{
			Name:    "Run tests",
			Type:    config.Run,
			Command: fmt.Sprintf("%s run test:ci", packageManager),
		}}
	}

	if npmTaskDefined(ls, "test") {
		if hasJestLabel {
			return []config.Step{{
				Name:    "Run tests",
				Type:    config.Run,
				Command: fmt.Sprintf("%s run test --ci --runInBand --reporters=default --reporters=jest-junit", packageManager),
			}}
		}

		return []config.Step{{
			Name:    "Run tests",
			Type:    config.Run,
			Command: fmt.Sprintf("%s run test", packageManager),
		}}
	}

	if hasJestLabel {
		return []config.Step{{
			Type:    config.Run,
			Name:    "Run tests with Jest",
			Command: "jest --ci --runInBand --reporters=default --reporters=jest-junit",
		}}
	}

	return []config.Step{{
		Name:    "Run tests",
		Type:    config.Run,
		Command: fmt.Sprintf("%s test", packageManager),
	}}
}

func nodeTestJob(ls labels.LabelSet) *Job {
	hasJestLabel := ls[labels.TestJest].Valid

	packageManager := "npm"
	if ls[labels.PackageManagerYarn].Valid {
		packageManager = "yarn"
	}

	steps := []config.Step{
		checkoutStep(ls[labels.DepsNode]),
	}

	if ls[labels.DepsNode].HasLockFile {
		steps = append(steps, config.Step{
			Type:    config.OrbCommand,
			Command: "node/install-packages",
			Parameters: config.OrbCommandParameters{
				"pkg-manager": packageManager,
			}})
	} else {
		steps = append(steps, config.Step{
			Comment: "Update the default install command as the project doesn't have a lock file",
			Type:    config.OrbCommand,
			Command: "node/install-packages",
			Parameters: config.OrbCommandParameters{
				"cache-path":          "~/project/node_modules",
				"override-ci-command": fmt.Sprintf("%s install", packageManager),
			}})
	}

	if hasJestLabel && ls[labels.DepsNode].Dependencies["jest-junit"] == "" {
		if packageManager == "yarn" {
			steps = append(steps, config.Step{
				Type:    config.Run,
				Command: "yarn add jest-junit --ignore-workspace-root-check",
			})
		} else {
			steps = append(steps, config.Step{
				Type:    config.Run,
				Command: "npm install jest-junit",
			})
		}
	}
	steps = append(steps, nodeTestSteps(ls, packageManager)...)

	if hasJestLabel {
		steps = append(steps, config.Step{
			Type:    config.OrbCommand,
			Command: "store_test_results",
			Parameters: config.OrbCommandParameters{
				"path": "./test-results/",
			},
		})
	}

	job := config.Job{
		Name:             "test-node",
		Comment:          "Install node dependencies and run tests",
		Executor:         "node/default",
		WorkingDirectory: workingDirectory(ls[labels.DepsNode]),
		Steps:            steps}

	if hasJestLabel {
		job.Environment = map[string]string{
			"JEST_JUNIT_OUTPUT_DIR": "./test-results/",
		}
	}

	return &Job{
		Job:  job,
		Type: TestJob,
		Orbs: map[string]string{"node": nodeOrb},
	}
}

func GenerateNodeJobs(ls labels.LabelSet) []*Job {
	if !ls[labels.DepsNode].Valid {
		return nil
	}

	return []*Job{nodeTestJob(ls)}
}

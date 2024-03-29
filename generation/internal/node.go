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

func nodePackageManager(ls labels.LabelSet) string {
	if ls[labels.PackageManagerYarn].Valid {
		return "yarn"
	}
	return "npm"
}

func nodeRunCommand(ls labels.LabelSet, task string) string {
	if task == "test" {
		// don't fail for default test value for new node projects
		if ls[labels.DepsNode].Tasks[task] == "echo \"Error: no test specified\" && exit 1" {
			return "echo \\\"No test specified in package.json\\\""
		}
		return fmt.Sprintf("%s test --passWithNoTests", nodePackageManager(ls))
	}
	return fmt.Sprintf("%s run %s", nodePackageManager(ls), task)
}

func nodeInitialSteps(ls labels.LabelSet) []config.Step {
	steps := []config.Step{
		checkoutStep(ls[labels.DepsNode]),
	}

	installParams := config.OrbCommandParameters{
		"pkg-manager": nodePackageManager(ls),
	}
	if !ls[labels.DepsNode].HasLockFile {
		installParams = config.OrbCommandParameters{
			"cache-path":          "~/project/node_modules",
			"override-ci-command": fmt.Sprintf("%s install", nodePackageManager(ls)),
		}
	}

	steps = append(steps,
		config.Step{
			Type:       config.OrbCommand,
			Command:    "node/install-packages",
			Parameters: installParams,
		},
	)

	return steps
}

func nodeTestSteps(ls labels.LabelSet) []config.Step {
	hasJestLabel := ls[labels.TestJest].Valid

	if npmTaskDefined(ls, "test:ci") {
		return []config.Step{{
			Name:    "Run tests",
			Type:    config.Run,
			Command: nodeRunCommand(ls, "test:ci"),
		}}
	}

	if npmTaskDefined(ls, "test") {
		if hasJestLabel {
			return []config.Step{{
				Name: "Run tests",
				Type: config.Run,
				Command: fmt.Sprintf(
					"%s run test --ci --runInBand --reporters=default --reporters=jest-junit",
					nodePackageManager(ls)),
			}}
		}

		return []config.Step{{
			Name:    "Run tests",
			Type:    config.Run,
			Command: nodeRunCommand(ls, "test"),
		}}
	}

	if npmTaskDefined(ls, "test:unit") {
		return []config.Step{{
			Name:    "Run tests",
			Type:    config.Run,
			Command: nodeRunCommand(ls, "test:unit"),
		}}
	}

	if hasJestLabel {
		return []config.Step{{
			Type:    config.Run,
			Name:    "Run tests with Jest",
			Command: "./node_modules/.bin/jest --ci --runInBand --reporters=default --reporters=jest-junit",
		}}
	}

	return []config.Step{}
}

func nodeTestJob(ls labels.LabelSet) *Job {
	hasJestLabel := ls[labels.TestJest].Valid

	steps := nodeInitialSteps(ls)

	if hasJestLabel && ls[labels.DepsNode].Dependencies["jest-junit"] == "" {
		if nodePackageManager(ls) == "yarn" {
			command := "yarn add jest-junit --ignore-workspace-root-check"

			if ls[labels.PackageManagerYarn].Version == "berry" {
				// yarn-berry doesn't support --ignore-workspace-root-check and it's not needed in this case
				command = "yarn add jest-junit"
			}
			steps = append(steps, config.Step{
				Type:    config.Run,
				Command: command,
			})
		} else {
			steps = append(steps, config.Step{
				Type:    config.Run,
				Command: "npm install jest-junit",
			})
		}
	}

	testSteps := nodeTestSteps(ls)
	if len(testSteps) == 0 {
		return nil
	}
	steps = append(steps, testSteps...)

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

func nodeBuildJob(ls labels.LabelSet) *Job {
	// Possible build task names in order of preference
	buildTasks := []string{
		"build:ci",
		"build:production",
		"build:prod",
		"build",
		"build:development",
		"build:dev",
	}

	steps := nodeInitialSteps(ls)

	for _, task := range buildTasks {
		if npmTaskDefined(ls, task) {

			steps = append(steps, []config.Step{
				{
					Type:    config.Run,
					Command: nodeRunCommand(ls, task),
				},
				createArtifactsDirStep,
				{
					Type:    config.Run,
					Comment: "Copy output to artifacts dir",
					Name:    "Copy artifacts",
					Command: "cp -R build dist public .output .next .docusaurus ~/artifacts 2>/dev/null || true",
				},
				storeArtifactsStep("node-build")}...)

			return &Job{
				Job: config.Job{
					Name:             "build-node",
					Comment:          "Build node project",
					Executor:         "node/default",
					WorkingDirectory: workingDirectory(ls[labels.DepsNode]),
					Steps:            steps,
				},
				Type: ArtifactJob,
				Orbs: map[string]string{"node": nodeOrb},
			}
		}
	}

	return nil
}

func GenerateNodeJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.DepsNode].Valid {
		return nil
	}

	testJob := nodeTestJob(ls)
	if testJob != nil {
		jobs = append(jobs, testJob)
	}

	buildJob := nodeBuildJob(ls)
	if buildJob != nil {
		jobs = append(jobs, buildJob)
	}

	return jobs
}

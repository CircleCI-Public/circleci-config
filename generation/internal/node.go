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

	if npmTaskDefined(ls, "test:ci") {
		return []config.Step{{
			Type:    config.Run,
			Command: fmt.Sprintf("%s run test:ci", packageManager),
		}}
	}

	if npmTaskDefined(ls, "test") {
		return []config.Step{{
			Type:    config.Run,
			Command: fmt.Sprintf("%s run test", packageManager),
		}}
	}

	if ls[labels.TestJest].Valid {
		return []config.Step{{
			Type:    config.Run,
			Command: "jest",
		}}
	}

	return []config.Step{{
		Type:    config.Run,
		Command: fmt.Sprintf("%s test", packageManager),
	}}
}

func nodeTestJob(ls labels.LabelSet) *Job {
	steps := initialSteps(ls[labels.DepsNode])

	packageManager := "npm"
	if ls[labels.PackageManagerYarn].Valid {
		packageManager = "yarn"
	}

	steps = append(steps, config.Step{
		Type:    config.OrbCommand,
		Command: "node/install-packages",
		Parameters: config.OrbCommandParameters{
			"pkg-manager": packageManager,
		},
	})
	steps = append(steps, nodeTestSteps(ls, packageManager)...)

	return &Job{
		Job: config.Job{Name: "test-node",
			Comment:  "Install node dependencies and run tests",
			Executor: "node/default",
			Steps:    steps},
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

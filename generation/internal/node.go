package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const nodeOrb = "circleci/node@5"

func npmTaskDefined(ls labels.LabelSet, task string) bool {
	return ls[labels.DepsNode].Tasks[task] != ""
}

func nodeTestSteps(ls labels.LabelSet) []config.Step {
	if ls[labels.TestJest].Valid {
		return []config.Step{{
			Type:    config.Run,
			Command: "jest",
		}}
	}
	if npmTaskDefined(ls, "test:ci") {
		return []config.Step{{
			Type:    config.Run,
			Command: "npm run test:ci",
		}}
	}

	return []config.Step{{
		Type:    config.Run,
		Command: "npm test",
	}}
}

func nodeTestJob(ls labels.LabelSet) *Job {
	steps := initialSteps(ls[labels.DepsNode])

	steps = append(steps, config.Step{
		Type:    config.OrbCommand,
		Command: "node/install-packages",
	})
	steps = append(steps, nodeTestSteps(ls)...)

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

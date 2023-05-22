package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const nodeOrb = "circleci/node@5"

func npmTaskDefined(matches labels.MatchSet, task string) bool {
	return matches[labels.DepsNode].Tasks[task] != ""
}

func nodeTestSteps(matches labels.MatchSet) []config.Step {
	if matches[labels.TestJest].Valid {
		return []config.Step{{
			Type:    config.Run,
			Command: "jest",
		}}
	}
	if npmTaskDefined(matches, "test:ci") {
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

func nodeTestJob(matches labels.MatchSet) *Job {
	steps := initialSteps(matches[labels.DepsNode])

	steps = append(steps, config.Step{
		Type:    config.OrbCommand,
		Command: "node/install-packages",
	})
	steps = append(steps, nodeTestSteps(matches)...)

	return &Job{
		Job: config.Job{Name: "test-node",
			Comment:  "Install node dependencies and run tests",
			Executor: "node/default",
			Steps:    steps},
		Type: TestJob,
		Orbs: map[string]string{"node": nodeOrb},
	}
}

func GenerateNodeJobs(matches labels.MatchSet) []*Job {
	if !matches[labels.DepsNode].Valid {
		return nil
	}
	return []*Job{nodeTestJob(matches)}
}

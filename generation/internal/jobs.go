package internal

import (
	"fmt"
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

type Type uint32

// This type is to help us build a workflow graph
const (
	TestJob Type = iota // generic test job
	ArtifactJob
)

type Job struct {
	config.Job
	Type
	// map of orb name (e.g. "slack") to registry key (e.g. "circleci/slack@4.12.5")
	// for orbs required by this job
	Orbs map[string]string
}

func BuildConfig(matches labels.MatchSet, jobs []*Job) config.Config {
	if len(jobs) == 0 {
		return fallbackConfig
	}

	// Can jobs not just be "cast" to []*config.Jobs somehow?
	configJobs := make([]*config.Job, len(jobs))
	for i := range jobs {
		configJobs[i] = &jobs[i].Job
	}

	workflows := buildWorkflows(jobs)

	return config.Config{
		Comment:   fmt.Sprintf("This config was automatically generated from your source code"),
		Workflows: workflows,
		Jobs:      configJobs,
		Orbs:      buildOrbs(jobs),
	}
}

var fallbackConfig = config.Config{
	Comment: "Couldn't automatically generate a config from your source code.\n" +
		"This is generic template to serve as a base for your custom config",
	// FIXME
}

func buildOrbs(jobs []*Job) []config.Orb {
	// merge all jobs orb maps
	orbsByName := make(map[string]string)
	for _, j := range jobs {
		for name, registryKey := range j.Orbs {
			orbsByName[name] = registryKey
		}
	}

	orbs := make([]config.Orb, len(orbsByName))
	i := 0
	for name, registryKey := range orbsByName {
		orbs[i] = config.Orb{Name: name, RegistryKey: registryKey}
		i++
	}

	return orbs
}

func buildWorkflows(jobs []*Job) []*config.Workflow {
	// For now, just generate one Workflow with all jobs, no "requires"
	// We might want to do something a bit smarter in the future
	workflowJobs := make([]config.WorkflowJob, len(jobs))
	for i, j := range jobs {
		workflowJobs[i].Job = &j.Job
	}

	return []*config.Workflow{{
		Name: "ci",
		Jobs: workflowJobs,
	}}
}

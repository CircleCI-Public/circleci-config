package internal

import (
	"fmt"
	"sort"

	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

type Type uint32

// This type is to help us build a workflow graph
const (
	TestJob Type = iota // generic test job
	ArtifactJob
	DeployJob
)

type Job struct {
	config.Job
	Type
	// map of orb name (e.g. "slack") to registry key (e.g. "circleci/slack@4.12.5")
	// for orbs required by this job
	Orbs map[string]string
}

func BuildConfig(ls labels.LabelSet, jobs []*Job) config.Config {
	if len(jobs) == 0 {
		return fallbackConfig
	}

	jobs = addStubJobs(jobs)

	// Can jobs not just be "cast" to []*config.Jobs somehow?
	configJobs := make([]*config.Job, len(jobs))
	for i := range jobs {
		configJobs[i] = &jobs[i].Job
	}

	workflows := buildWorkflows(jobs)

	return config.Config{
		Comment: fmt.Sprintf("This config was automatically generated from your source code\n"+
			"Stacks detected: %s", ls.String()),
		Workflows: workflows,
		Jobs:      configJobs,
		Orbs:      buildOrbs(jobs),
	}
}

func addStubJobs(jobs []*Job) []*Job {
	jobTypesPresent := map[Type]bool{}

	for _, j := range jobs {
		jobTypesPresent[j.Type] = true
	}

	if !jobTypesPresent[TestJob] {
		jobs = append(jobs, &stubTestJob)
	}
	if !jobTypesPresent[DeployJob] {
		jobs = append(jobs, &stubDeployJob)
	}

	return jobs
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

	// sort to make the slice order dependable for tests
	sort.Slice(orbs, func(i, j int) bool {
		return orbs[i].Name < orbs[j].Name
	})

	return orbs
}

func buildWorkflows(jobs []*Job) []*config.Workflow {
	// DeployJobs are added, but commented out
	workflowJobs := make([]config.WorkflowJob, len(jobs))
	for i, j := range jobs {
		workflowJobs[i].Job = &j.Job
		workflowJobs[i].CommentedOut = j.Type == DeployJob
		workflowJobs[i].Requires = workflowJobRequires(j, jobs)
	}

	return []*config.Workflow{{
		Name: "ci",
		Jobs: workflowJobs,
	}}
}

func workflowJobRequires(job *Job, allJobs []*Job) []*config.Job {
	jobsByType := getJobsByType(allJobs)

	if job.Type == ArtifactJob {
		return jobsByType[TestJob]
	}

	if job.Type == DeployJob {
		requires := []*config.Job{}
		requires = append(requires, jobsByType[TestJob]...)
		requires = append(requires, jobsByType[ArtifactJob]...)

		return requires
	}

	return []*config.Job{}
}

func getJobsByType(jobs []*Job) map[Type][]*config.Job {
	jobsByType := map[Type][]*config.Job{}

	for _, j := range jobs {
		jobsByType[j.Type] = append(jobsByType[j.Type], &j.Job)
	}

	return jobsByType
}

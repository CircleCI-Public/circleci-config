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
		return buildFallbackConfig(ls)
	}

	jobs = addStubJobs(ls, jobs)

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

func buildDeployJob(ls labels.LabelSet) *Job {
	deployJob := stubDeployJob
	deployJob.Steps = stubDeployJob.Steps
	deployJob.Steps = append(deployJob.Steps, getCICDSteps(ls)...)
	if emptyRepoLabel, ok := ls[labels.EmptyRepo]; ok && emptyRepoLabel.Valid {
		deployJob.Steps = append(deployJob.Steps, getEmptyJobSteps(ls)...)
	}

	return &deployJob
}

func buildFallbackConfig(ls labels.LabelSet) config.Config {
	deployJob := buildDeployJob(ls)
	deployJob.Comment = ""

	comment := `Couldn't automatically generate a config from your source code.
This is a generic template to serve as a base for your custom config
See: https://circleci.com/docs/configuration-reference`
	if len(ls) > 0 {
		comment = fmt.Sprintf("%s\n"+
			"Stacks detected: %s", comment, ls.String())

	}

	return config.Config{
		Comment: comment,
		Jobs: []*config.Job{
			&stubTestJob.Job,
			&stubArtifactJob.Job,
			&deployJob.Job,
		},
		Workflows: []*config.Workflow{
			{
				Name: "example",
				Jobs: []config.WorkflowJob{
					{
						Job: &stubTestJob.Job,
					}, {
						Job: &stubArtifactJob.Job,
						Requires: []*config.Job{
							&stubTestJob.Job,
						},
					}, {
						Job: &deployJob.Job,
						Requires: []*config.Job{
							&stubTestJob.Job,
						},
					},
				},
			},
		},
	}

}

func getCICDSteps(ls labels.LabelSet) []config.Step {
	steps := make([]config.Step, 0)

	cicdStepMap := map[string]string{
		labels.CICDGithubActions:  "github actions",
		labels.CICDGitlabWorkflow: "gitlab workflows",
		labels.CICDJenkins:        "jenkins",
	}

	for label, cicdName := range cicdStepMap {
		if ciLabel, ok := ls[label]; ok && ciLabel.Valid {
			steps = append(steps, config.Step{
				Type:    config.Run,
				Name:    fmt.Sprintf("found %s config", cicdName),
				Command: ":", // dont do anything just return status 0 and continue
			})
		}

	}

	return steps
}

func addStubJobs(ls labels.LabelSet, jobs []*Job) []*Job {
	jobTypesPresent := map[Type]bool{}

	for _, j := range jobs {
		jobTypesPresent[j.Type] = true
	}

	if !jobTypesPresent[TestJob] && !jobTypesPresent[ArtifactJob] {
		jobs = append(jobs, &stubTestJob)
	}
	if !jobTypesPresent[DeployJob] {
		deployJob := buildDeployJob(ls)
		jobs = append(jobs, deployJob)
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

	name := "build"

	for _, j := range jobs {
		if j.Type == TestJob {
			name = "build-and-test"
		}
	}

	return []*config.Workflow{{
		Name: name,
		Jobs: workflowJobs,
	}}
}

func workflowJobRequires(job *Job, allJobs []*Job) []*config.Job {
	jobsByType := getJobsByType(allJobs)

	if job.Type == ArtifactJob {
		return jobsByType[TestJob]
	}

	if job.Type == DeployJob {
		// if there are artifact jobs, it's sufficient to depend on them,
		// as they already depend on any test jobs themselves
		if len(jobsByType[ArtifactJob]) > 0 {
			return jobsByType[ArtifactJob]
		}
		return jobsByType[TestJob]
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

func getEmptyJobSteps(ls labels.LabelSet) []config.Step {
	steps := make([]config.Step, 0)
	steps = append(steps, config.Step{
		Type:    config.Run,
		Name:    "found empty repo",
		Command: ":", // dont do anything just return status 0 and continue
	})
	return steps
}

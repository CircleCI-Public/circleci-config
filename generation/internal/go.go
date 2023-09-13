package internal

import (
	"fmt"

	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

const privateModInstructions = `echo "go mod download will fail if you have private repositories 
One way to fix this for private go modules that are hosted in github:
  1. Add a GITHUB_TOKEN and GITHUB_USER to an org context. Please refer to https://circleci.com/docs/contexts/ for more informaiton on how to use contexts.
  2. Add a circleci config or use this generated config as a starting template
  3. Set the jobs to use the context that includes github token and user  
  4. Before downloading the modules you will need to add a step to execute \"go env -w GOPRIVATE=github.com/<OrgNameHere>\". 
	This lets allows go mod to install private repos under OrgNameHere.
  5. You will also need to run \"git config --global url.\"https://$GITHUB_USER:$GITHUB_TOKEN@github.com/<OrgNameHere>/\".insteadOf \"https://github.com/<OrgNameHere>/\"\"
  6. Finally include the \"go mod download\", it should be able to fetch your private libraries now. 
For gitlab private go modules follow the same instructions as above but include your GITLAB_TOKEN, GITLAB_USER.
Then use gitlab.com instead of github.com in steps 4 and 5.
Please take a look in https://go.dev/ref/mod#private-modules for more details.
"	
`

func goInitialSteps(ls labels.LabelSet) []config.Step {
	depsLabel := ls[labels.DepsGo]
	steps := []config.Step{
		checkoutStep(depsLabel),
	}
	if !depsLabel.HasLockFile {
		return steps
	}

	const goCacheKey = `go-mod-{{ checksum "go.sum" }}`
	return append(steps,
		config.Step{
			Type:     config.RestoreCache,
			CacheKey: goCacheKey,
		},
		config.Step{
			Type:    config.Run,
			Name:    "Download Go modules",
			Command: "go mod download",
		},
		config.Step{
			Type:    config.Run,
			Name:    "Print go mod help instructions",
			Command: privateModInstructions,
			When:    config.WhenTypeOnFail,
		}, 
		config.Step{
			Type:     config.SaveCache,
			CacheKey: goCacheKey,
			Path:     "/home/circleci/go/pkg/mod",
		},
	)
}

func goTestJob(ls labels.LabelSet) *Job {
	steps := goInitialSteps(ls)

	steps = append(steps, []config.Step{{
		Type:    config.Run,
		Name:    "Run tests",
		Command: "gotestsum --junitfile junit.xml",
	}, {
		Type: config.StoreTestResults,
		Path: "junit.xml",
	}}...)

	return &Job{
		Job: config.Job{
			Name:             "test-go",
			Comment:          "Install go modules and run tests",
			DockerImages:     []string{"cimg/go:1.20"},
			WorkingDirectory: workingDirectory(ls[labels.DepsGo]),
			Steps:            steps,
		},
		Type: TestJob,
	}
}

func goBuildJob(ls labels.LabelSet) *Job {
	steps := goInitialSteps(ls)

	steps = append(steps,
		createArtifactsDirStep,
		config.Step{
			Type:    config.Run,
			Name:    "Build executables",
			Command: fmt.Sprintf("go build -o %s ./...", artifactsPath),
		},
		storeArtifactsStep("executables"))

	return &Job{
		Job: config.Job{
			Name:         "build-go-executables",
			Comment:      "Build go executables and store them as artifacts",
			DockerImages: []string{"cimg/go:1.20"},
			Steps:        steps,
		},
		Type: ArtifactJob,
	}
}

func GenerateGoJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.DepsGo].Valid {
		return nil
	}

	jobs = append(jobs, goTestJob(ls))

	if ls[labels.ArtifactGoExecutable].Valid {
		jobs = append(jobs, goBuildJob(ls))
	}

	return jobs
}

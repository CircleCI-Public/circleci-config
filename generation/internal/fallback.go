package internal

import (
	"github.com/CircleCI-Public/circleci-config/config"
)

var stubTestJob = Job{
	Job: config.Job{
		Name:         "test",
		Comment:      "",
		DockerImages: []string{"cimg/base:stable"},
		Steps: []config.Step{
			{
				Type: config.Checkout,
			}, {
				Name:    "Run tests",
				Type:    config.Run,
				Comment: "Replace this with a real test runner invocation",
				Command: "echo 'replace me with real tests!' && false",
			},
		},
	},
	Type: TestJob,
}

var stubArtifactJob = Job{
	Job: config.Job{
		Name:         "build",
		Comment:      "",
		DockerImages: []string{"cimg/base:stable"},
		Steps: []config.Step{
			{
				Type: config.Checkout,
			}, {
				Type:    config.Run,
				Name:    "Build an artifact",
				Comment: "Replace this with steps to build a package, or executable",
				Command: "touch example.txt",
			}, {
				Type: config.StoreArtifacts,
				Path: "example.txt",
			},
		},
	},
	Type: ArtifactJob,
}

var stubDeployJob = Job{
	Job: config.Job{
		Name:         "deploy",
		Comment:      "This is an example deploy job, not actually used by the workflow",
		DockerImages: []string{"cimg/base:stable"},
		Steps: []config.Step{{
			Type:    config.Run,
			Name:    "deploy",
			Comment: "Replace this with steps to deploy to users",
			Command: "#e.g. ./deploy.sh",
		}},
	},
	Type: DeployJob,
}

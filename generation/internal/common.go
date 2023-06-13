package internal

import (
	"fmt"
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"github.com/alessio/shellescape"
)

var checkoutStep = config.Step{Type: config.Checkout}

// initialSteps returns a checkout step and, if necessary cd step
func initialSteps(depsLabel labels.Label) []config.Step {
	steps := []config.Step{checkoutStep}

	basePath := depsLabel.BasePath
	if basePath != "." {
		steps = append(steps, config.Step{
			Type:    config.Run,
			Name:    fmt.Sprintf("Change into '%s' directory", basePath),
			Command: fmt.Sprintf("cd '%s'", shellescape.Quote(basePath)),
		})
	}

	return steps
}

func withOrbAppDir(parameters config.OrbCommandParameters, depsLabel labels.Label) config.OrbCommandParameters {
	basePath := depsLabel.BasePath
	if basePath != "" && basePath != "." {
		parameters["app-dir"] = basePath
	}

	return parameters
}

const artifactsPath = "~/artifacts"

var createArtifactsDirStep = config.Step{
	Type:    config.Run,
	Name:    fmt.Sprintf("Create the %s directory if it doesn't exist", artifactsPath),
	Command: fmt.Sprintf("mkdir -p %s", artifactsPath),
}

var storeArtifactsStep = config.Step{
	Type: config.StoreArtifacts,
	Path: artifactsPath,
}

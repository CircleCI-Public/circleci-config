package internal

import (
	"fmt"
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"path/filepath"
)

const defaultCheckoutDir = "~/project"

func checkoutStep(depsLabel labels.Label) config.Step {
	if depsLabel.BasePath == "." {
		return config.Step{Type: config.Checkout}
	}
	return config.Step{
		Type: config.Checkout,
		Path: defaultCheckoutDir,
	}
}

func workingDirectory(depsLabel labels.Label) string {
	if depsLabel.BasePath == "." {
		return "."
	}
	return filepath.Join(defaultCheckoutDir, depsLabel.BasePath)
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

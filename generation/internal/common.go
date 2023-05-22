package internal

import (
	"fmt"
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"github.com/alessio/shellescape"
)

var checkoutStep = config.Step{Type: config.Checkout}

// initialSteps returns a checkout step and, if necessary cd step
func initialSteps(depsMatch labels.Match) []config.Step {
	steps := []config.Step{checkoutStep}

	basePath := depsMatch.BasePath
	if basePath != "." {
		steps = append(steps, config.Step{
			Type:    config.Run,
			Name:    fmt.Sprintf("Change into '%s' directory", basePath),
			Command: fmt.Sprintf("cd '%s'", shellescape.Quote(basePath)),
		})
	}

	return steps
}

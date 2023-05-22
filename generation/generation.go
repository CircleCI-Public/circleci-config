package generation

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/generation/internal"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func GenerateConfig(matches labels.MatchSet) config.Config {
	var generatedJobs []*internal.Job
	generatedJobs = append(generatedJobs, internal.GenerateNodeJobs(matches)...)
	generatedJobs = append(generatedJobs, internal.GenerateGoJobs(matches)...)
	return internal.BuildConfig(matches, generatedJobs)
}

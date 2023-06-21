package generation

import (
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/generation/internal"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func GenerateConfig(labels labels.LabelSet) config.Config {
	var generatedJobs []*internal.Job
	generatedJobs = append(generatedJobs, internal.GenerateNodeJobs(labels)...)
	generatedJobs = append(generatedJobs, internal.GenerateGoJobs(labels)...)
	generatedJobs = append(generatedJobs, internal.GeneratePythonJobs(labels)...)
	generatedJobs = append(generatedJobs, internal.GenerateRubyJobs(labels)...)
	generatedJobs = append(generatedJobs, internal.GenerateRustJobs(labels)...)
	return internal.BuildConfig(labels, generatedJobs)
}

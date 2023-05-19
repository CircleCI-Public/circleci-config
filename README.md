# circleci-config

Go library to work with CircleCI config files.

## Features

Currently, only implements:

### Rules for labeling codebases

The [labeling package](labeling) implements rules for detecting the tech stack used in a
codebase:

```go
c := codebase.LocalCodebase{}
matches := labeling.ApplyAllRules(c)
// matches: map of labels like "deps:node" to a Match structure containing more details
```

Rules for different stacks can be found in the [labeling/internal](labeling/internal) directory.

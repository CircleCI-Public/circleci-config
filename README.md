# circleci-config

Go library to work with CircleCI config files.

## Features

Currently, only implements feature required for config inference:

### Rules for labeling codebases

The [labeling package](labeling) implements rules for detecting the tech stack used in a
codebase:

```go
c := codebase.LocalCodebase{}
matches := labeling.ApplyAllRules(c)
// matches: map of labels like "deps:node" to a Match structure containing more details
```

Rules for different stacks can be found in the [internal](labeling/internal) directory.

### Generating jobs for a given set of labels

The [generation package](generation) takes a set of label matches and produces CI jobs for them,
and then assembles them into a config:

```go
config := generation.GenerateConfig(matches)
// config: data structure that represents a CircleCI config with workflows, jobs, orbs, etc.
```

### Config serialization to YAML

The [config package](config) defines structs that represent a CircleCI config and that can
be serialized to YAML.

See [the TestConfig_YamlNode test](config/config_test.go) for an example.

```go
yamlNode := config.YamlNode() // yamlNode: a gopkg.in/yaml.v3 yaml.Node
```

Not all possible configs can be represented, only the ones needed for inference.

## Adding a new language or software stack

Adding support for a new stack consists of three parts:

1. Defining [labels](labeling/labels/labels.go) to identify the stack and its variants. Add
   `"deps:..."` labels for dependency management, `"test:..."` labels for test runners, and
   `"build:..."` labels for build systems (if they are different).
2. Implementing rules that tag codebases with the above labels in the
   [labeling/internal directory](labeling/internal). Create a new file for each language.
   Then add those rules to the [`ApplyAllRules` function](labeling/labeling.go).
3. Implement a function that given those rules generates jobs in the
   [generation/internal directory](generation/internal). Again, create a new file for each language.
   Add that function to the list of calls in [`GenerateConfig`](generation/generation.go).

Of course, also add tests for the new rules and config generation code.

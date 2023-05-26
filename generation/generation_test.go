package generation

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling"
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"gopkg.in/yaml.v3"
)

func mustEncode(node *yaml.Node) string {
	buf := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buf)
	// Just to make it more compact to write expected results
	encoder.SetIndent(2)
	err := encoder.Encode(node)
	if err != nil {
		fmt.Print(err)
		panic("error encoding yaml")
	}
	return buf.String()
}

func testEncode(t *testing.T, node config.Node, expected string) {
	yamlStr := mustEncode(node.YamlNode())
	if yamlStr != expected {
		t.Errorf("\ngot:     %q\nexpected:%q", yamlStr, expected)
	}
}

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		testName string
		labels   labels.LabelSet
		expected string
	}{
		{
			testName: "node and go codebases in subdirs",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "node-dir"},
				},
				labels.DepsGo: labels.Label{
					Key:       labels.DepsGo,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "go-dir"},
				}},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:go:go-dir,deps:node:node-dir
version: 2.1
orbs:
  node: circleci/node@5
jobs:
  test-node:
    # Install node dependencies and run tests
    executor: node/default
    steps:
      - checkout
      - run:
          name: Change into 'node-dir' directory
          command: cd 'node-dir'
      - node/install-packages:
          pkg-manager: npm
      - run:
          command: npm test
  test-go:
    # Install go modules, run go vet and tests
    docker:
      - image: cimg/go:1.20
    steps:
      - checkout
      - run:
          name: Change into 'go-dir' directory
          command: cd 'go-dir'
      - restore_cache:
          key: go-mod-{{ checksum "go.sum" }}
      - run:
          name: Download Go modules
          command: go mod download
      - save_cache:
          key: go-mod-{{ checksum "go.sum" }}
          paths:
            - /home/circleci/go/pkg/mod
      - run:
          name: Run go vet
          command: go vet ./...
      - run:
          name: Run tests
          command: gotestsum --junitfile junit.xml
      - store_test_results:
          path: junit.xml
workflows:
  ci:
    jobs:
      - test-node
      - test-go
`,
		},
		{
			testName: "node codebase with yarn.lock",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
				labels.PackageManagerYarn: labels.Label{
					Key:       labels.PackageManagerYarn,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:node:.,package_manager:yarn:.
version: 2.1
orbs:
  node: circleci/node@5
jobs:
  test-node:
    # Install node dependencies and run tests
    executor: node/default
    steps:
      - checkout
      - node/install-packages:
          pkg-manager: yarn
      - run:
          command: yarn test
workflows:
  ci:
    jobs:
      - test-node
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gotConfig := GenerateConfig(tt.labels)
			testEncode(t, gotConfig, tt.expected)
		})
	}
}

func TestDogfood(t *testing.T) {
	expectedConfig := `# This config was automatically generated from your source code
# Stacks detected: artifact:go-executable:,deps:go:.
version: 2.1
jobs:
  test-go:
    # Install go modules, run go vet and tests
    docker:
      - image: cimg/go:1.20
    steps:
      - checkout
      - restore_cache:
          key: go-mod-{{ checksum "go.sum" }}
      - run:
          name: Download Go modules
          command: go mod download
      - save_cache:
          key: go-mod-{{ checksum "go.sum" }}
          paths:
            - /home/circleci/go/pkg/mod
      - run:
          name: Run go vet
          command: go vet ./...
      - run:
          name: Run tests
          command: gotestsum --junitfile junit.xml
      - store_test_results:
          path: junit.xml
  build-go-executables:
    # Build go executables and store them as artifacts
    docker:
      - image: cimg/go:1.20
    steps:
      - checkout
      - restore_cache:
          key: go-mod-{{ checksum "go.sum" }}
      - run:
          name: Download Go modules
          command: go mod download
      - save_cache:
          key: go-mod-{{ checksum "go.sum" }}
          paths:
            - /home/circleci/go/pkg/mod
      - run:
          name: Create the ~/artifacts directory if it doesn't exist
          command: mkdir -p ~/artifacts
      - run:
          name: Build executables
          command: go build -o ~/artifacts ./...
      - store_artifacts:
          path: ~/artifacts
workflows:
  ci:
    jobs:
      - test-go
      - build-go-executables
`

	c := codebase.LocalCodebase{BasePath: ".."}
	ls := labeling.ApplyAllRules(c)
	gotConfig := GenerateConfig(ls)
	testEncode(t, gotConfig, expectedConfig)
}

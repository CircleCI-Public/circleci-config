package generation

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/CircleCI-Public/circleci-config/config"
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
	d := cmp.Diff(expected, yamlStr)
	if d != "" {
		t.Errorf("\ngot:     %q\nexpected:%q", yamlStr, expected)
		fmt.Printf("diff: %s", d)
	}
}

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		testName string
		labels   labels.LabelSet
		expected string
	}{
		{
			testName: "no labels generates fallback config",
			labels:   labels.LabelSet{},
			expected: `# Couldn't automatically generate a config from your source code.
# This is generic template to serve as a base for your custom config
# See: https://circleci.com/docs/configuration-reference
version: 2.1
jobs:
  test:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      # Replace this with a real test runner invocation
      - run:
          name: Run tests
          command: echo 'replace me with real tests!' && false
  build:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      # Replace this with steps to build a package, or executable
      - run:
          name: Build an artifact
          command: touch example.txt
      - store_artifacts:
          path: example.txt
  deploy:
    # This is an example deploy job, not actually used by the workflow
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
workflows:
  example:
    jobs:
      - test
      - build:
          requires:
            - test
      - deploy:
          requires:
            - test
`,
		}, {
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
          name: Run tests
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
  deploy:
    # This is an example deploy job, not actually used by the workflow
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
workflows:
  ci:
    jobs:
      - test-node
      - test-go
    # - deploy
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
          name: Run tests
          command: yarn test
  deploy:
    # This is an example deploy job, not actually used by the workflow
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
workflows:
  ci:
    jobs:
      - test-node
    # - deploy
`,
		},
		{
			testName: "node codebase with jest tests",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
				labels.TestJest: labels.Label{
					Key:       labels.TestJest,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:node:.,test:jest:.
version: 2.1
orbs:
  node: circleci/node@5
jobs:
  test-node:
    # Install node dependencies and run tests
    executor: node/default
    environment:
      JEST_JUNIT_OUTPUT_DIR: ./test-results/
    steps:
      - checkout
      - node/install-packages:
          pkg-manager: npm
      - run:
          command: npm install jest-junit
      - run:
          name: Run tests with Jest
          command: jest --ci --runInBand --reporters=default --reporters=jest-junit
      - store_test_results:
          path: ./test-results/
  deploy:
    # This is an example deploy job, not actually used by the workflow
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
workflows:
  ci:
    jobs:
      - test-node
    # - deploy
`,
		},
		{

			testName: "python codebase with poetry package manager",
			labels: labels.LabelSet{
				labels.DepsPython: labels.Label{
					Key:       labels.DepsPython,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
				labels.PackageManagerPoetry: labels.Label{
					Key:       labels.PackageManagerPoetry,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:python:.,package_manager:poetry:.
version: 2.1
orbs:
  python: circleci/python@2
jobs:
  test-python:
    # Install dependencies and run tests
    executor: python/default
    steps:
      - checkout
      - python/install-packages:
          pkg-manager: poetry
      - run:
          name: Run tests
          command: poetry run pytest --junitxml=junit.xml
      - store_test_results:
          path: junit.xml
  deploy:
    # This is an example deploy job, not actually used by the workflow
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
workflows:
  ci:
    jobs:
      - test-python
    # - deploy
`,
		},
		{

			testName: "python codebase in a subdirectory with poetry package manager",
			labels: labels.LabelSet{
				labels.DepsPython: labels.Label{
					Key:       labels.DepsPython,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "x"},
				},
				labels.PackageManagerPoetry: labels.Label{
					Key:       labels.PackageManagerPoetry,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "x"},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:python:x,package_manager:poetry:x
version: 2.1
orbs:
  python: circleci/python@2
jobs:
  test-python:
    # Install dependencies and run tests
    executor: python/default
    steps:
      - checkout
      - run:
          name: Change into 'x' directory
          command: cd 'x'
      - python/install-packages:
          pkg-manager: poetry
      - run:
          name: Run tests
          command: poetry run pytest --junitxml=junit.xml
      - store_test_results:
          path: junit.xml
  deploy:
    # This is an example deploy job, not actually used by the workflow
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
workflows:
  ci:
    jobs:
      - test-python
    # - deploy
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

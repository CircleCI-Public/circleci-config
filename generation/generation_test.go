package generation

import (
	"bytes"
	"fmt"
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling"
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"gopkg.in/yaml.v3"
	"testing"
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
			expected: "# This config was automatically generated from your source code\n" +
				"# Stacks detected: deps:go:go-dir,deps:node:node-dir\n" +
				"version: 2.1\n" +
				"orbs:\n" +
				"  node: circleci/node@5\n" +
				"jobs:\n" +
				"  test-node:\n" +
				"    # Install node dependencies and run tests\n" +
				"    executor: node/default\n" +
				"    steps:\n" +
				"      - checkout\n" +
				"      - run:\n" +
				"          name: Change into 'node-dir' directory\n" +
				"          command: cd 'node-dir'\n" +
				"      - node/install-packages\n" +
				"      - run:\n" +
				"          command: npm test\n" +
				"  test-go:\n" +
				"    # Install go modules, run go vet and tests\n" +
				"    docker:\n" +
				"      - image: cimg/go\n" +
				"    steps:\n" +
				"      - checkout\n" +
				"      - run:\n" +
				"          name: Change into 'go-dir' directory\n" +
				"          command: cd 'go-dir'\n" +
				"      - restore_cache:\n" +
				"          key: go-mod-{{ checksum \"go.sum\" }}\n" +
				"      - run:\n" +
				"          name: Download Go modules\n" +
				"          command: go mod download\n" +
				"      - save_cache:\n" +
				"          key: go-mod-{{ checksum \"go.sum\" }}\n" +
				"          paths:\n" +
				"            - /home/circleci/go/pkg/mod\n" +
				"      - run:\n" +
				"          name: Run go vet\n" +
				"          command: go vet ./...\n" +
				"      - run:\n" +
				"          name: Run tests\n" +
				"          command: gotestsum --junitfile junit.xml\n" +
				"      - store_test_results:\n" +
				"          path: junit.xml\n" +
				"workflows:\n" +
				"  ci:\n" +
				"    jobs:\n" +
				"      - test-node\n" +
				"      - test-go\n",
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
	expectedConfig := "# This config was automatically generated from your source code\n# Stacks detected: deps:go:.\nversion: 2.1\njobs:\n  test-go:\n    # Install go modules, run go vet and tests\n    docker:\n      - image: cimg/go\n    steps:\n      - checkout\n      - restore_cache:\n          key: go-mod-{{ checksum \"go.sum\" }}\n      - run:\n          name: Download Go modules\n          command: go mod download\n      - save_cache:\n          key: go-mod-{{ checksum \"go.sum\" }}\n          paths:\n            - /home/circleci/go/pkg/mod\n      - run:\n          name: Run go vet\n          command: go vet ./...\n      - run:\n          name: Run tests\n          command: gotestsum --junitfile junit.xml\n      - store_test_results:\n          path: junit.xml\nworkflows:\n  ci:\n    jobs:\n      - test-go\n"

	c := codebase.LocalCodebase{BasePath: ".."}
	ls := labeling.ApplyAllRules(c)
	gotConfig := GenerateConfig(ls)
	testEncode(t, gotConfig, expectedConfig)
}

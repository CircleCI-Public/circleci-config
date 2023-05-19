package config

import (
	"bytes"
	"fmt"
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

func testEncode(t *testing.T, node Node, expected string) {
	yamlStr := mustEncode(node.YamlNode())
	if yamlStr != expected {
		t.Errorf("\ngot:     %q\nexpected:%q", yamlStr, expected)
	}
}

func TestConfig_YamlNode(t *testing.T) {
	var nodeTestJob = Job{
		Name:        "node-test-job",
		DockerImage: "cimg/base",
		Steps: []Step{{
			Type: Checkout,
		}, {
			Type:     RestoreCache,
			CacheKey: "npm-cache-key",
		}, {
			Type:    Run,
			Command: "npm install",
		}, {
			Type:     SaveCache,
			CacheKey: "npm-cache-key",
			Path:     "./node_modules",
		}, {
			Type:    Run,
			Command: "npm test",
		}},
	}

	var npmBuildJob = Job{
		Name:        "node-build-job",
		DockerImage: "cimg/base",
		Steps: []Step{{
			Type: Checkout,
		}, {
			Type:     RestoreCache,
			CacheKey: "npm-cache-key",
		}, {
			Type:    Run,
			Command: "npm pack",
		}},
	}

	tests := []struct {
		testName string
		config   Config
		expected string
	}{
		{
			testName: "config",
			config: Config{
				Orbs: []Orb{{Name: "node", RegistryKey: "circleci/node@5"}},
				Jobs: []*Job{&nodeTestJob, &npmBuildJob},
				Workflows: []*Workflow{{
					Name: "node-workflow",
					Jobs: []WorkflowJob{
						{
							Job: &nodeTestJob,
						}, {
							Job:      &npmBuildJob,
							Requires: []*Job{&nodeTestJob},
						}},
				}},
			},
			expected: "version: 2.1\n" +
				"orbs:\n" +
				"  node: circleci/node@5\n" +
				"jobs:\n" +
				"  node-test-job:\n" +
				"    docker:\n" +
				"      - image: cimg/base\n" +
				"    steps:\n" +
				"      - checkout\n" +
				"      - restore_cache:\n" +
				"          key: npm-cache-key\n" +
				"      - run:\n" +
				"          command: npm install\n" +
				"      - save_cache:\n" +
				"          key: npm-cache-key\n" +
				"          paths:\n" +
				"            - ./node_modules\n" +
				"      - run:\n" +
				"          command: npm test\n" +
				"  node-build-job:\n" +
				"    docker:\n" +
				"      - image: cimg/base\n" +
				"    steps:\n" +
				"      - checkout\n" +
				"      - restore_cache:\n" +
				"          key: npm-cache-key\n" +
				"      - run:\n" +
				"          command: npm pack\n" +
				"workflows:\n" +
				"  node-workflow:\n" +
				"    jobs:\n" +
				"      - node-test-job\n" +
				"      - node-build-job:\n" +
				"          requires:\n" +
				"            - node-test-job\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			testEncode(t, tt.config, tt.expected)
		})
	}
}

func TestWorkflowJob_YamlNode(t *testing.T) {
	var job1 = Job{Name: "job1"}
	var job2 = Job{Name: "job2"}
	var job3 = Job{Name: "job3"}

	tests := []struct {
		testName string
		job      *Job
		requires []*Job
		expected string
	}{
		{
			testName: "no requires",
			job:      &job1,
			expected: "job1\n",
		}, {
			testName: "empty requires",
			job:      &job1,
			requires: []*Job{},
			expected: "job1\n",
		}, {
			testName: "2 requires",
			job:      &job1,
			requires: []*Job{&job2, &job3},
			expected: "job1:\n  requires:\n    - job2\n    - job3\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			wj := WorkflowJob{
				Job:      tt.job,
				Requires: tt.requires,
			}
			testEncode(t, wj, tt.expected)
		})
	}
}

func TestJob_YamlNode(t *testing.T) {
	tests := []struct {
		testName string
		job      Job
		expected string
	}{
		{
			testName: "job with docker image",
			job: Job{
				Name:        "job",
				Comment:     "This is a job that uses docker",
				DockerImage: "cimg/base",
				Steps: []Step{{
					Type: Checkout,
				}, {
					Type:    Run,
					Comment: "get deps",
					Command: "npm install",
				}},
			},
			expected: "# This is a job that uses docker\n" +
				"docker:\n" +
				"  - image: cimg/base\n" +
				"steps:\n" +
				"  - checkout\n" +
				"  # get deps\n" +
				"  - run:\n" +
				"      command: npm install\n",
		}, {
			testName: "job with executor",
			job: Job{
				Name:     "job",
				Executor: "x",
			},
			expected: "executor: x\nsteps: []\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			testEncode(t, tt.job, tt.expected)
		})
	}
}

func TestStep_YamlNode(t *testing.T) {
	tests := []struct {
		testName string
		step     Step
		expected string
	}{
		{
			testName: "checkout",
			step: Step{
				Type: Checkout,
			},
			expected: "checkout\n",
		}, {
			testName: "checkout with comment",
			step: Step{
				Type:    Checkout,
				Comment: "first, checkout the code",
			},
			expected: "# first, checkout the code\ncheckout\n",
		}, {
			testName: "run without name",
			step: Step{
				Type:    Run,
				Command: "echo Hi",
			},
			expected: "run:\n  command: echo Hi\n",
		}, {
			testName: "run with name",
			step: Step{
				Type:    Run,
				Name:    "Say Hi",
				Command: "echo Hi",
			},
			expected: "run:\n  name: Say Hi\n  command: echo Hi\n",
		}, {
			testName: "run with name and comment",
			step: Step{
				Type:    Run,
				Name:    "Say Hi",
				Comment: "greet",
				Command: "echo Hi",
			},
			expected: "# greet\nrun:\n  name: Say Hi\n  command: echo Hi\n",
		}, {
			testName: "save_cache",
			step: Step{
				Type:     SaveCache,
				CacheKey: "cache-key",
				Path:     "/stuff",
			},
			expected: "save_cache:\n  key: cache-key\n  paths:\n    - /stuff\n",
		}, {
			testName: "restore_cache",
			step: Step{
				Type:     RestoreCache,
				CacheKey: "cache-key",
			},
			expected: "restore_cache:\n  key: cache-key\n",
		}, {
			testName: "store_artifacts",
			step: Step{
				Type: StoreArtifacts,
				Path: "/out",
			},
			expected: "store_artifacts:\n  path: /out\n",
		}, {
			testName: "store_test_results",
			step: Step{
				Type: StoreTestResults,
				Path: "/test-results",
			},
			expected: "store_test_results:\n  path: /test-results\n",
		}, {
			testName: "orb command without params",
			step: Step{
				Type:    OrbCommand,
				Command: "orb/do_something",
			},
			expected: "orb/do_something\n",
		}, {
			testName: "orb command with params",
			step: Step{
				Type:       OrbCommand,
				Command:    "orb/do_something",
				Parameters: OrbCommandParameters{"x": "1", "y": "2"},
			},
			expected: "orb/do_something:\n  x: 1\n  y: 2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			testEncode(t, tt.step, tt.expected)
		})
	}
}

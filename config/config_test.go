package config

import (
	"testing"
)

func testEncode(t *testing.T, node Node, expected string) {
	yamlStr := yamlNodeToString(node.YamlNode())
	if yamlStr != expected {
		t.Errorf("\ngot:     %q\nexpected:%q", yamlStr, expected)
	}
}

func TestConfig_String(t *testing.T) {
	var nodeTestJob = Job{
		Name:         "node-test-job",
		DockerImages: []string{"cimg/base"},
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
		Name:         "node-build-job",
		DockerImages: []string{"cimg/base"},
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
			got := tt.config.String()
			if got != tt.expected {
				t.Errorf("\ngot:     %q\nexpected:%q", got, tt.expected)
			}
		})
	}
}

func TestWorkflow_YamlNode(t *testing.T) {
	var job1 = Job{Name: "job1"}
	var job2 = Job{Name: "job2"}
	var job3 = Job{Name: "job3"}

	tests := []struct {
		testName string
		workflow Workflow
		expected string
	}{
		{
			testName: "no jobs",
			workflow: Workflow{
				Name: "w",
			},
			expected: "jobs: []\n",
		}, {
			testName: "3 jobs independent",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job: &job1,
					}, {
						Job: &job2,
					}, {
						Job: &job3,
					},
				},
			},
			expected: "jobs:\n  - job1\n  - job2\n  - job3\n",
		}, {
			testName: "3 jobs sequential",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job: &job1,
					}, {
						Job:      &job2,
						Requires: []*Job{&job1},
					}, {
						Job:      &job3,
						Requires: []*Job{&job2},
					},
				},
			},
			expected: `jobs:
  - job1
  - job2:
      requires:
        - job1
  - job3:
      requires:
        - job2
`,
		}, {
			testName: "3 jobs fan-out",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job: &job1,
					}, {
						Job:      &job2,
						Requires: []*Job{&job1},
					}, {
						Job:      &job3,
						Requires: []*Job{&job1},
					},
				},
			},
			expected: `jobs:
  - job1
  - job2:
      requires:
        - job1
  - job3:
      requires:
        - job1
`,
		}, {
			testName: "3 jobs fan-in",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job: &job1,
					}, {
						Job: &job2,
					}, {
						Job:      &job3,
						Requires: []*Job{&job1, &job2},
					},
				},
			},
			expected: `jobs:
  - job1
  - job2
  - job3:
      requires:
        - job1
        - job2
`,
		}, {
			testName: "3 jobs, job1 commented out",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job:          &job1,
						CommentedOut: true,
					}, {
						Job: &job2,
					}, {
						Job: &job3,
					},
				},
			},
			expected: `jobs:
  # - job1
  - job2
  - job3
`,
		}, {
			testName: "3 jobs, job2 (with requires) commented out",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job: &job1,
					}, {
						Job:          &job2,
						CommentedOut: true,
						Requires:     []*Job{&job1},
					}, {
						Job: &job3,
					},
				},
			},
			expected: `jobs:
  - job1
  # - job2:
  #     requires:
  #       - job1
  - job3
`,
		}, {
			testName: "3 jobs, job3 (with requires) commented out",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job: &job1,
					}, {
						Job: &job2,
					}, {
						Job:          &job3,
						CommentedOut: true,
						Requires:     []*Job{&job1, &job2},
					},
				},
			},
			// Note that when the last job is commented out the indentation is different
			expected: `jobs:
  - job1
  - job2
# - job3:
#     requires:
#       - job1
#       - job2
`,
		}, {
			testName: "3 jobs, job1 and job2 commented out",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job:          &job1,
						CommentedOut: true,
					}, {
						Job:          &job2,
						CommentedOut: true,
					}, {
						Job: &job3,
					},
				},
			},
			expected: `jobs:
  # - job1
  # - job2
  - job3
`,
		}, {
			testName: "3 jobs, job1 and job3 commented out",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job:          &job1,
						CommentedOut: true,
					}, {
						Job: &job2,
					}, {
						Job:          &job3,
						CommentedOut: true,
					},
				},
			},
			// Note that when the last job is commented out the indentation is different
			expected: `jobs:
  # - job1
  - job2
# - job3
`,
		}, {
			testName: "3 jobs, job2 and job3 commented out",
			workflow: Workflow{
				Name: "w",
				Jobs: []WorkflowJob{
					{
						Job: &job1,
					}, {
						Job:          &job2,
						CommentedOut: true,
					}, {
						Job:          &job3,
						CommentedOut: true,
					},
				},
			},
			// Note that when the last jobs are commented out the indentation is different
			expected: `jobs:
  - job1
# - job2
# - job3
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			testEncode(t, tt.workflow, tt.expected)
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
				Name:         "job",
				Comment:      "This is a job that uses docker",
				DockerImages: []string{"cimg/base"},
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
		{
			testName: "job with docker image and environment variables",
			job: Job{
				Name:         "job",
				Comment:      "This is a job that uses docker",
				DockerImages: []string{"cimg/base"},
				Environment: map[string]string{
					"FOO": "bar",
					"BAZ": "qux",
				},
				Steps: []Step{{
					Type: Checkout,
				}, {
					Type:    Run,
					Comment: "get deps",
					Command: "npm install",
				}},
			},
			expected: `# This is a job that uses docker
docker:
  - image: cimg/base
environment:
  BAZ: qux
  FOO: bar
steps:
  - checkout
  # get deps
  - run:
      command: npm install
`,
		},
		{
			testName: "job with executor and working dir",
			job: Job{
				Name:             "job",
				Executor:         "x",
				WorkingDirectory: "dir",
			},
			expected: "executor: x\nworking_directory: dir\nsteps: []\n",
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
			testName: "checkout with path",
			step: Step{
				Type: Checkout,
				Path: "subdir",
			},
			expected: "checkout:\n  path: subdir\n",
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

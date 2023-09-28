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
# This is a generic template to serve as a base for your custom config
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
		},

		{
			testName: "cicd included in fallback config",
			labels: labels.LabelSet{
				labels.CICDGithubActions: labels.Label{
					Key:   labels.CICDGithubActions,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
			},
			expected: `# Couldn't automatically generate a config from your source code.
# This is a generic template to serve as a base for your custom config
# See: https://circleci.com/docs/configuration-reference
# Stacks detected: cicd:github-actions:.
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
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
      - run:
          name: found github actions config
          command: ':'
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
		},
		{
			testName: "node and go codebases in subdirs",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "node-dir", HasLockFile: true, Tasks: map[string]string{"test": "false"}},
				},
				labels.DepsGo: labels.Label{
					Key:       labels.DepsGo,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "go-dir", HasLockFile: true},
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
    working_directory: ~/project/node-dir
    steps:
      - checkout:
          path: ~/project
      - node/install-packages:
          pkg-manager: npm
      - run:
          name: Print node install help instructions
          command: |-
            echo "One cause for node package install failure is if you have private repositories that it can't reach
            One way to fix this for private npm packages:
              1. Use the npm CLI's \"login\" command to create a token (usually saved in your user's \"~/.npmrc\" file)
                For more info, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=set%20the%20%24npm_token%20environment%20variable%20in%20circleci
              2. Add a NPM_TOKEN to an org context
                For info on how to use contexts, see https://circleci.com/docs/contexts/
              3. Add a .circleci/config.yml to your repository or use this config.yml as a starting template
              4. Configure the jobs to use the context that includes NPM_TOKEN
              5. Add a step to inject your NPM_TOKEN environment variable into npm before \"install-packages\"
                For an example, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=the%20deploy%20job%20has%20several%20steps%20that%20run%20to%20authenticate%20with%20and%20publish%20to"
          when: on_fail
      - run:
          name: Run tests
          command: npm test --passWithNoTests
  test-go:
    # Install go modules and run tests
    docker:
      - image: cimg/go:1.20
    working_directory: ~/project/go-dir
    steps:
      - checkout:
          path: ~/project
      - restore_cache:
          key: go-mod-{{ checksum "go.sum" }}
      - run:
          name: Download Go modules
          command: go mod download
      - run:
          name: Print go mod help instructions
          command: "echo \"go mod download will fail if you have private repositories \nOne way to fix this for private go modules that are hosted in github:\n  1. Add a GITHUB_TOKEN and GITHUB_USER to an org context. Please refer to https://circleci.com/docs/contexts/ for more informaiton on how to use contexts.\n  2. Add a .circleci/config.yml to your repository or use this config.yml as a starting template\n  3. Configure the jobs to use the newly created context which includes GITHUB_TOKEN and GITHUB_USER  \n  4. Before downloading the modules you will need to add a step to execute \\\"go env -w GOPRIVATE=github.com/<OrgNameHere>\\\". \n\tThis allows go mod to install private repos under OrgNameHere.\n  5. You will also need to run \\\"git config --global url.\\\"https://$GITHUB_USER:$GITHUB_TOKEN@github.com/<OrgNameHere>/\\\".insteadOf \\\"https://github.com/<OrgNameHere>/\\\"\\\"\n  6. Finally include the \\\"go mod download\\\" it should be able to fetch your private libraries now. \nFor gitlab private go modules, follow the same instructions as above but include your GITLAB_TOKEN and GITLAB_USER.\nThen use gitlab.com instead of github.com in steps 4 and 5.\nSee https://go.dev/ref/mod#private-modules for more details.\""
          when: on_fail
      - save_cache:
          key: go-mod-{{ checksum "go.sum" }}
          paths:
            - /home/circleci/go/pkg/mod
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
  build-and-test:
    jobs:
      - test-node
      - test-go
    # - deploy:
    #     requires:
    #       - test-node
    #       - test-go
`,
		},
		{
			testName: "node codebase with yarn.lock",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: ".", HasLockFile: true, Tasks: map[string]string{"test": "false"}},
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
          name: Print node install help instructions
          command: |-
            echo "One cause for node package install failure is if you have private repositories that it can't reach
            One way to fix this for private npm packages:
              1. Use the npm CLI's \"login\" command to create a token (usually saved in your user's \"~/.npmrc\" file)
                For more info, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=set%20the%20%24npm_token%20environment%20variable%20in%20circleci
              2. Add a NPM_TOKEN to an org context
                For info on how to use contexts, see https://circleci.com/docs/contexts/
              3. Add a .circleci/config.yml to your repository or use this config.yml as a starting template
              4. Configure the jobs to use the context that includes NPM_TOKEN
              5. Add a step to inject your NPM_TOKEN environment variable into npm before \"install-packages\"
                For an example, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=the%20deploy%20job%20has%20several%20steps%20that%20run%20to%20authenticate%20with%20and%20publish%20to"
          when: on_fail
      - run:
          name: Run tests
          command: yarn test --passWithNoTests
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
  build-and-test:
    jobs:
      - test-node
    # - deploy:
    #     requires:
    #       - test-node
`,
		},
		{
			testName: "node codebase with yarn.lock and .yarnrc.yml file",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: ".", HasLockFile: true},
				},
				labels.TestJest: labels.Label{
					Key:       labels.TestJest,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
				labels.PackageManagerYarn: labels.Label{
					Key:       labels.PackageManagerYarn,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: ".", Version: "berry"},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:node:.,package_manager:yarn:.,test:jest:.
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
          pkg-manager: yarn
      - run:
          name: Print node install help instructions
          command: |-
            echo "One cause for node package install failure is if you have private repositories that it can't reach
            One way to fix this for private npm packages:
              1. Use the npm CLI's \"login\" command to create a token (usually saved in your user's \"~/.npmrc\" file)
                For more info, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=set%20the%20%24npm_token%20environment%20variable%20in%20circleci
              2. Add a NPM_TOKEN to an org context
                For info on how to use contexts, see https://circleci.com/docs/contexts/
              3. Add a .circleci/config.yml to your repository or use this config.yml as a starting template
              4. Configure the jobs to use the context that includes NPM_TOKEN
              5. Add a step to inject your NPM_TOKEN environment variable into npm before \"install-packages\"
                For an example, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=the%20deploy%20job%20has%20several%20steps%20that%20run%20to%20authenticate%20with%20and%20publish%20to"
          when: on_fail
      - run:
          command: yarn add jest-junit
      - run:
          name: Run tests with Jest
          command: ./node_modules/.bin/jest --ci --runInBand --reporters=default --reporters=jest-junit
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
  build-and-test:
    jobs:
      - test-node
    # - deploy:
    #     requires:
    #       - test-node
`,
		},
		{
			testName: "node codebase without a lock file",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: ".", HasLockFile: false, Tasks: map[string]string{"test": "false"}},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:node:.
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
          cache-path: ~/project/node_modules
          override-ci-command: npm install
      - run:
          name: Print node install help instructions
          command: |-
            echo "One cause for node package install failure is if you have private repositories that it can't reach
            One way to fix this for private npm packages:
              1. Use the npm CLI's \"login\" command to create a token (usually saved in your user's \"~/.npmrc\" file)
                For more info, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=set%20the%20%24npm_token%20environment%20variable%20in%20circleci
              2. Add a NPM_TOKEN to an org context
                For info on how to use contexts, see https://circleci.com/docs/contexts/
              3. Add a .circleci/config.yml to your repository or use this config.yml as a starting template
              4. Configure the jobs to use the context that includes NPM_TOKEN
              5. Add a step to inject your NPM_TOKEN environment variable into npm before \"install-packages\"
                For an example, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=the%20deploy%20job%20has%20several%20steps%20that%20run%20to%20authenticate%20with%20and%20publish%20to"
          when: on_fail
      - run:
          name: Run tests
          command: npm test --passWithNoTests
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
  build-and-test:
    jobs:
      - test-node
    # - deploy:
    #     requires:
    #       - test-node
`,
		},
		{
			testName: "node codebase with jest tests",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: ".", HasLockFile: true},
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
          name: Print node install help instructions
          command: |-
            echo "One cause for node package install failure is if you have private repositories that it can't reach
            One way to fix this for private npm packages:
              1. Use the npm CLI's \"login\" command to create a token (usually saved in your user's \"~/.npmrc\" file)
                For more info, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=set%20the%20%24npm_token%20environment%20variable%20in%20circleci
              2. Add a NPM_TOKEN to an org context
                For info on how to use contexts, see https://circleci.com/docs/contexts/
              3. Add a .circleci/config.yml to your repository or use this config.yml as a starting template
              4. Configure the jobs to use the context that includes NPM_TOKEN
              5. Add a step to inject your NPM_TOKEN environment variable into npm before \"install-packages\"
                For an example, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=the%20deploy%20job%20has%20several%20steps%20that%20run%20to%20authenticate%20with%20and%20publish%20to"
          when: on_fail
      - run:
          command: npm install jest-junit
      - run:
          name: Run tests with Jest
          command: ./node_modules/.bin/jest --ci --runInBand --reporters=default --reporters=jest-junit
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
  build-and-test:
    jobs:
      - test-node
    # - deploy:
    #     requires:
    #       - test-node
`,
		},
		{
			testName: "node codebase with build task",
			labels: labels.LabelSet{
				labels.DepsNode: labels.Label{
					Key:       labels.DepsNode,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: ".", HasLockFile: true, Tasks: map[string]string{"build": "echo hi"}},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:node:.
version: 2.1
orbs:
  node: circleci/node@5
jobs:
  build-node:
    # Build node project
    executor: node/default
    steps:
      - checkout
      - node/install-packages:
          pkg-manager: npm
      - run:
          name: Print node install help instructions
          command: |-
            echo "One cause for node package install failure is if you have private repositories that it can't reach
            One way to fix this for private npm packages:
              1. Use the npm CLI's \"login\" command to create a token (usually saved in your user's \"~/.npmrc\" file)
                For more info, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=set%20the%20%24npm_token%20environment%20variable%20in%20circleci
              2. Add a NPM_TOKEN to an org context
                For info on how to use contexts, see https://circleci.com/docs/contexts/
              3. Add a .circleci/config.yml to your repository or use this config.yml as a starting template
              4. Configure the jobs to use the context that includes NPM_TOKEN
              5. Add a step to inject your NPM_TOKEN environment variable into npm before \"install-packages\"
                For an example, see https://circleci.com/blog/publishing-npm-packages-using-circleci-2-0/#:~:text=the%20deploy%20job%20has%20several%20steps%20that%20run%20to%20authenticate%20with%20and%20publish%20to"
          when: on_fail
      - run:
          command: npm run build
      - run:
          name: Create the ~/artifacts directory if it doesn't exist
          command: mkdir -p ~/artifacts
      # Copy output to artifacts dir
      - run:
          name: Copy artifacts
          command: cp -R build dist public .output .next .docusaurus ~/artifacts 2>/dev/null || true
      - store_artifacts:
          path: ~/artifacts
          destination: node-build
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
  build:
    jobs:
      - build-node
    # - deploy:
    #     requires:
    #       - build-node
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
    docker:
      - image: cimg/python:3.8-node
    steps:
      - checkout
      - python/install-packages:
          pkg-manager: poetry
      - run:
          name: Run tests
          command: poetry run pytest --junitxml=junit.xml || ((($? == 5)) && echo 'Did not find any tests to run.')
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
  build-and-test:
    jobs:
      - test-python
    # - deploy:
    #     requires:
    #       - test-python
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
    docker:
      - image: cimg/python:3.8-node
    working_directory: ~/project/x
    steps:
      - checkout:
          path: ~/project
      - python/install-packages:
          pkg-manager: poetry
      - run:
          name: Run tests
          command: poetry run pytest --junitxml=junit.xml || ((($? == 5)) && echo 'Did not find any tests to run.')
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
  build-and-test:
    jobs:
      - test-python
    # - deploy:
    #     requires:
    #       - test-python
`,
		},
		{
			testName: "python project with a manage.py file using poetry",
			labels: labels.LabelSet{
				labels.DepsPython: labels.Label{
					Key:   labels.DepsPython,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
				labels.FileManagePy: labels.Label{
					Key:       labels.FileManagePy,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:python:.,file:manage.py:.
version: 2.1
orbs:
  python: circleci/python@2
jobs:
  test-python:
    # Install dependencies and run tests
    docker:
      - image: cimg/python:3.8-node
    steps:
      - checkout
      - python/install-packages
      - run:
          name: Run tests
          command: python manage.py test
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
  build-and-test:
    jobs:
      - test-python
    # - deploy:
    #     requires:
    #       - test-python
`,
		},
		{
			testName: "python project with a manage.py file using pipenv",
			labels: labels.LabelSet{
				labels.DepsPython: labels.Label{
					Key:   labels.DepsPython,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
				labels.FileManagePy: labels.Label{
					Key:       labels.FileManagePy,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
				labels.PackageManagerPipenv: labels.Label{
					Key:       labels.PackageManagerPipenv,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:python:.,file:manage.py:.,package_manager:pipenv:.
version: 2.1
orbs:
  python: circleci/python@2
jobs:
  test-python:
    # Install dependencies and run tests
    docker:
      - image: cimg/python:3.8-node
    steps:
      - checkout
      - python/install-packages:
          args: --dev
          pkg-manager: pipenv
      - run:
          name: Run tests
          command: pipenv run python manage.py test
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
  build-and-test:
    jobs:
      - test-python
    # - deploy:
    #     requires:
    #       - test-python
`,
		},
		{
			testName: "python project with a manage.py file using poetry",
			labels: labels.LabelSet{
				labels.DepsPython: labels.Label{
					Key:   labels.DepsPython,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
				labels.FileManagePy: labels.Label{
					Key:       labels.FileManagePy,
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
# Stacks detected: deps:python:.,file:manage.py:.,package_manager:poetry:.
version: 2.1
orbs:
  python: circleci/python@2
jobs:
  test-python:
    # Install dependencies and run tests
    docker:
      - image: cimg/python:3.8-node
    steps:
      - checkout
      - python/install-packages:
          pkg-manager: poetry
      - run:
          name: Run tests
          command: poetry run python manage.py test
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
  build-and-test:
    jobs:
      - test-python
    # - deploy:
    #     requires:
    #       - test-python
`,
		},
		{
			testName: "python project with a setup.py and tox",
			labels: labels.LabelSet{
				labels.DepsPython: labels.Label{
					Key:   labels.DepsPython,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
				labels.FileSetupPy: labels.Label{
					Key:       labels.FileSetupPy,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
				labels.TestTox: labels.Label{
					Key:       labels.TestTox,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:python:.,file:setup.py:.,test:tox:.
version: 2.1
orbs:
  python: circleci/python@2
jobs:
  test-python:
    # Install dependencies and run tests
    docker:
      - image: cimg/python:3.8-node
    steps:
      - checkout
      - python/install-packages:
          pkg-manager: pip-dist
      - python/install-packages:
          args: tox
          pkg-manager: pip-dist
      - run:
          name: Run tests
          command: tox
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
  build-and-test:
    jobs:
      - test-python
    # - deploy:
    #     requires:
    #       - test-python
`,
		},
		{
			testName: "python project with a .python-version file",
			labels: labels.LabelSet{
				labels.DepsPython: labels.Label{
					Key:   labels.DepsPython,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"python": "3.1.1",
						},
					},
				},
				labels.PackageManagerPipenv: labels.Label{
					Key:       labels.PackageManagerPipenv,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
				labels.CICDGitlabWorkflow: labels.Label{
					Key:       labels.CICDGitlabWorkflow,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "."},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: cicd:gitlab-workflows:.,deps:python:.,package_manager:pipenv:.
version: 2.1
orbs:
  python: circleci/python@2
jobs:
  test-python:
    # Install dependencies and run tests
    docker:
      - image: cimg/python:3.1.1-node
    steps:
      - checkout
      - python/install-packages:
          args: --dev
          pkg-manager: pipenv
      - run:
          name: Run tests
          command: pipenv run pytest --junitxml=junit.xml || ((($? == 5)) && echo 'Did not find any tests to run.')
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
      - run:
          name: found gitlab workflows config
          command: ':'
workflows:
  build-and-test:
    jobs:
      - test-python
    # - deploy:
    #     requires:
    #       - test-python
`,
		},
		{
			testName: "java project based using maven",
			labels: labels.LabelSet{
				labels.DepsJava: labels.Label{
					Key:   labels.DepsJava,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
				labels.CICDJenkins: labels.Label{
					Key:   labels.CICDJenkins,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: cicd:jenkins:.,deps:java:.
version: 2.1
jobs:
  test-java:
    docker:
      - image: cimg/openjdk:17.0
    steps:
      - checkout
      - run:
          name: Calculate cache key
          command: |-
            find . -name 'pom.xml' -o -name 'gradlew*' -o -name '*.gradle*' | \
                    sort | xargs cat > /tmp/CIRCLECI_CACHE_KEY
      - restore_cache:
          key: cache-{{ checksum "/tmp/CIRCLECI_CACHE_KEY" }}
      - run:
          command: mvn verify
      - store_test_results:
          path: target/surefire-reports
      - save_cache:
          key: cache-{{ checksum "/tmp/CIRCLECI_CACHE_KEY" }}
          paths:
            - ~/.m2/repository
  deploy:
    # This is an example deploy job, not actually used by the workflow
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
      - run:
          name: found jenkins config
          command: ':'
workflows:
  build-and-test:
    jobs:
      - test-java
    # - deploy:
    #     requires:
    #       - test-java
`,
		},
		{
			testName: "java project using gradle",
			labels: labels.LabelSet{
				labels.DepsJava: labels.Label{
					Key:   labels.DepsJava,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
					},
				},
				labels.ToolGradle: labels.Label{
					Key:   labels.ToolGradle,
					Valid: true,
				},
				labels.CICDGithubActions: labels.Label{
					Key:   labels.CICDGithubActions,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".github/workflows",
					},
				},
			},
			expected: `# This config was automatically generated from your source code
# Stacks detected: cicd:github-actions:.github/workflows,deps:java:.,tool:gradle:
version: 2.1
jobs:
  test-java:
    docker:
      - image: cimg/openjdk:17.0
    steps:
      - checkout
      - run:
          name: Calculate cache key
          command: |-
            find . -name 'pom.xml' -o -name 'gradlew*' -o -name '*.gradle*' | \
                    sort | xargs cat > /tmp/CIRCLECI_CACHE_KEY
      - restore_cache:
          key: cache-{{ checksum "/tmp/CIRCLECI_CACHE_KEY" }}
      - run:
          command: ./gradlew check
      - store_test_results:
          path: build/test-results
      - save_cache:
          key: cache-{{ checksum "/tmp/CIRCLECI_CACHE_KEY" }}
          paths:
            - ~/.gradle/caches
      - store_artifacts:
          path: build/reports
  deploy:
    # This is an example deploy job, not actually used by the workflow
    docker:
      - image: cimg/base:stable
    steps:
      # Replace this with steps to deploy to users
      - run:
          name: deploy
          command: '#e.g. ./deploy.sh'
      - run:
          name: found github actions config
          command: ':'
workflows:
  build-and-test:
    jobs:
      - test-java
    # - deploy:
    #     requires:
    #       - test-java
`,
		},

		{
			testName: "go codebase without lockfile",
			labels: labels.LabelSet{
				labels.DepsGo: labels.Label{
					Key:       labels.DepsGo,
					Valid:     true,
					LabelData: labels.LabelData{BasePath: "go-dir", HasLockFile: false},
				}},
			expected: `# This config was automatically generated from your source code
# Stacks detected: deps:go:go-dir
version: 2.1
jobs:
  test-go:
    # Install go modules and run tests
    docker:
      - image: cimg/go:1.20
    working_directory: ~/project/go-dir
    steps:
      - checkout:
          path: ~/project
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
  build-and-test:
    jobs:
      - test-go
    # - deploy:
    #     requires:
    #       - test-go
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

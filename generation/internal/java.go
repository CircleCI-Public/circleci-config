package internal

import (
	"fmt"
	"github.com/CircleCI-Public/circleci-config/config"
	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

// latest LTS version
const javaDockerImage = "cimg/openjdk:17.0"

func javaTestJob(ls labels.LabelSet) *Job {
	var cacheKey string
	var cachePath string
	var testCommand string
	var testResultsPath string
	var testReportsPath string

	if ls[labels.ToolGradle].Valid {
		gradleBuildFile := "build.gradle"
		if ls[labels.FileBuildGradleKts].Valid {
			gradleBuildFile = "build.gradle.kts"
		}
		cacheKey = fmt.Sprintf(`gradle-{{ checksum "%s" }}-{{ checksum "gradlew" }}`, gradleBuildFile)
		cachePath = "~/.gradle/caches"
		testCommand = "gradlew check"
		testResultsPath = "build/test-results"
		testReportsPath = "build/reports"
	} else {
		cacheKey = `maven-{{ checksum "pom.xml" }}`
		cachePath = "~/.m2/repository"
		testCommand = "mvn verify"
		testResultsPath = "target/surefire-reports"
	}

	steps := []config.Step{
		checkoutStep(ls[labels.DepsJava]),
		{
			Type:     config.RestoreCache,
			CacheKey: cacheKey,
		},
		{
			Type:    config.Run,
			Command: testCommand,
		},
		{
			Type: config.StoreTestResults,
			Path: testResultsPath,
		},
		{
			Type:     config.SaveCache,
			CacheKey: cacheKey,
			Path:     cachePath,
		},
	}

	if testReportsPath != "" {
		steps = append(steps, config.Step{
			Type: config.StoreArtifacts,
			Name: "Store test reports as artifacts",
			Path: testReportsPath,
		})
	}

	return &Job{
		Job: config.Job{
			Name:             "test-java",
			DockerImages:     []string{javaDockerImage},
			WorkingDirectory: workingDirectory(ls[labels.DepsJava]),
			Steps:            steps,
		},
		Type: TestJob,
	}
}

func GenerateJavaJobs(ls labels.LabelSet) (jobs []*Job) {
	if !ls[labels.DepsJava].Valid {
		return nil
	}

	return append(jobs, javaTestJob(ls))
}

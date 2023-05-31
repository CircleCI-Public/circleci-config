package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"testing"
)
import "github.com/go-git/go-git/v5"

func TestInferConfig(t *testing.T) {
	// Adding inference tests:
	// 1. Add the git url (https) to this list
	// 2. Add the expected config as a file in testdata/expected/REPO_NAME.yml
	tests := []struct {
		url string
	}{
		{url: "https://github.com/CircleCI-Public/circleci-demo-go"},
		{url: "https://github.com/CircleCI-Public/circleci-demo-react-native"},
		{url: "https://github.com/CircleCI-Public/circleci-demo-javascript-express"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Error(err)
				return
			}
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				t.Error(err)
				return
			}
			dir := fmt.Sprintf("%s/repos/%s", cacheDir, u.Path)
			_, err = git.PlainClone(dir, false, &git.CloneOptions{
				URL:          tt.url,
				SingleBranch: true,
				Depth:        1})
			if errors.Is(err, git.ErrRepositoryAlreadyExists) {
				fmt.Printf("Warning: Using existing cloned repo at %s\n", dir)
			} else if err != nil {
				t.Error(err)
				return
			}

			got := inferConfig(dir)
			expectedConfigFile := fmt.Sprintf("testdata/expected/%s.yml", path.Base(u.Path))
			expected, err := os.ReadFile(expectedConfigFile)
			if err != nil {
				t.Error(err)
				return
			}

			if got != string(expected) {
				t.Errorf("\ngot:\n%s\nexpected:\n%s", got, expected)
			}
		})
	}

}

package labeling

import (
	"testing"

	"github.com/CircleCI-Public/circleci-config/labeling/labels"
	"github.com/google/go-cmp/cmp"
)

func TestCodebase_ApplyPhpRules(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]string
		expectedLabels []labels.Label
	}{
		{
			name: "phpunit in composer file",
			files: map[string]string{
				"composer.json": composerWithPhpUnit,
			},
			expectedLabels: []labels.Label{
				{
					Key:   labels.DepsPhp,
					Valid: true,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"laravel/framework":       "5.4.*",
							"symfony/http-foundation": "3.4.35",
							"phpunit/phpunit":         "~4.0",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCodebase{tt.files}
			expected := make(labels.LabelSet)
			for _, label := range tt.expectedLabels {
				// all should be Valid
				label.Valid = true
				expected[label.Key] = label
			}
			got := ApplyAllRules(c)
			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("ApplyAllRules mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

const composerWithPhpUnit = `
{
  "name": "laravel/laravel",
  "description": "The Laravel Framework.",
  "keywords": ["framework", "laravel"],
  "license": "MIT",
  "type": "project",
  "require": {
    "laravel/framework": "5.4.*",
    "symfony/http-foundation": "3.4.35"
  },
  "require-dev": {
    "phpunit/phpunit": "~4.0"
  }
}			
`

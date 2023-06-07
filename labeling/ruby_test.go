package labeling

import (
	"reflect"
	"testing"

	"github.com/CircleCI-Public/circleci-config/labeling/labels"
)

func TestCodebase_ApplyRubyRules(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]string
		expectedLabels []labels.Label
	}{
		{
			name: "Ruby version",
			files: map[string]string{
				"Gemfile": rubyGemfile,
			},

			expectedLabels: []labels.Label{
				{
					Key: labels.DepsRuby,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"ruby": "2.7.8",
						},
					},
				},
			},
		},
		{
			name: "Ruby version w/ rspec, pg",
			files: map[string]string{
				"Gemfile": rubyGemfileWithRailsRSpec,
			},

			expectedLabels: []labels.Label{
				{
					Key: labels.DepsRuby,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"ruby":  "2.7.8",
							"rspec": "true",
							"pg":    "true",
						},
					},
				},
			},
		},
		{
			name: "Ruby version with engine info",
			files: map[string]string{
				"Gemfile": rubyGemfileWithEngine,
			},

			expectedLabels: []labels.Label{
				{
					Key: labels.DepsRuby,
					LabelData: labels.LabelData{
						BasePath: ".",
						Dependencies: map[string]string{
							"ruby": "1.9.3",
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

			if !reflect.DeepEqual(got, expected) {
				t.Errorf("\n"+
					"got      %v\n"+
					"expected %v",
					got,
					expected)
			}
		})
	}
}

const rubyGemfile = `
source 'https://rubygems.org'

git_source(:github) do |repo_name|
  repo_name = "#{repo_name}/#{repo_name}" unless repo_name.include?('/')
  "https://github.com/#{repo_name}.git"
end

ruby '2.7.8'

# Bundle edge Rails instead: gem 'rails', github: 'rails/rails'
gem 'rails', '~> 6.0.1'
`

const rubyGemfileWithRailsRSpec = `
source 'https://rubygems.org'

git_source(:github) do |repo_name|
  repo_name = "#{repo_name}/#{repo_name}" unless repo_name.include?('/')
  "https://github.com/#{repo_name}.git"
end

ruby '2.7.8'

# Bundle edge Rails instead: gem 'rails', github: 'rails/rails'
gem 'rails', '~> 6.0.1'

gem 'rspec-rails', '4.0.0.beta3'

# Use postgresql as the database for Active Record
gem 'pg', '~> 0.18'

`

const rubyGemfileWithEngine = `
source 'https://rubygems.org'

git_source(:github) do |repo_name|
  repo_name = "#{repo_name}/#{repo_name}" unless repo_name.include?('/')
  "https://github.com/#{repo_name}.git"
end

ruby '1.9.3', :engine => 'jruby', :engine_version => '1.6.7'

# Bundle edge Rails instead: gem 'rails', github: 'rails/rails'
gem 'rails', '~> 6.0.1'`
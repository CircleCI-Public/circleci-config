package config

import (
	"bytes"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

type Node interface {
	YamlNode() *yaml.Node
}

type Config struct {
	Comment   string
	Workflows []*Workflow
	Jobs      []*Job
	Orbs      []Orb
}

func (c Config) String() string {
	buf := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	err := encoder.Encode(c.YamlNode())
	if err != nil {
		return fmt.Sprintf("[Could not encode config: %v])", err)
	}
	return buf.String()
}

func (c Config) YamlNode() *yaml.Node {
	configNodes := []*yaml.Node{yScalar("version"), yScalar("2.1")}

	orbsYaml := make([]*yaml.Node, 2*len(c.Orbs))
	for i, o := range c.Orbs {
		orbsYaml[2*i] = yScalar(o.Name)
		orbsYaml[2*i+1] = o.YamlNode()
	}

	if len(orbsYaml) != 0 {
		configNodes = append(configNodes, yScalar("orbs"), yMap(orbsYaml...))
	}

	jobsYaml := make([]*yaml.Node, 2*len(c.Jobs))
	for i, j := range c.Jobs {
		jobsYaml[2*i] = yScalar(j.Name)
		jobsYaml[2*i+1] = j.YamlNode()
	}

	workflowsYaml := make([]*yaml.Node, 2*len(c.Workflows))
	for i, w := range c.Workflows {
		workflowsYaml[2*i] = yScalar(w.Name)
		workflowsYaml[2*i+1] = w.YamlNode()
	}

	configNodes = append(configNodes,
		yScalar("jobs"), yMap(jobsYaml...),
		yScalar("workflows"), yMap(workflowsYaml...))

	return yCommentedMap(c.Comment, configNodes...)
}

type Workflow struct {
	Name string
	Jobs []WorkflowJob
}

func (w Workflow) YamlNode() *yaml.Node {
	workflowJobsYaml := make([]*yaml.Node, len(w.Jobs))
	for i, j := range w.Jobs {
		workflowJobsYaml[i] = j.YamlNode()
	}

	return yMap(yScalar("jobs"), ySeq(workflowJobsYaml...))
}

type Orb struct {
	Name        string
	RegistryKey string
}

func (o Orb) YamlNode() *yaml.Node {
	return yScalar(o.RegistryKey)
}

// WorkflowJob are the references to the jobs that appear in the Workflow definitions
// For the actual job definitions (that appear under the top-level "jobs:" key) see Job type below
type WorkflowJob struct {
	Job      *Job
	Requires []*Job
}

func (wj WorkflowJob) YamlNode() *yaml.Node {
	nameYaml := yScalar(wj.Job.Name)
	if wj.Requires == nil || len(wj.Requires) == 0 {
		return nameYaml
	}
	requiresYaml := make([]*yaml.Node, len(wj.Requires))
	for i, r := range wj.Requires {
		requiresYaml[i] = yScalar(r.Name)
	}
	return yMap(nameYaml, yMap(
		yScalar("requires"), ySeq(requiresYaml...)))
}

// Job definitions as they appear under config top-level "jobs:" key
type Job struct {
	Name    string
	Comment string
	// The following two fields are mutually exclusive
	DockerImage string
	Executor    string
	Steps       []Step
	Environment map[string]string
}

func (j Job) YamlNode() *yaml.Node {
	stepsYaml := make([]*yaml.Node, len(j.Steps))
	for i, s := range j.Steps {
		stepsYaml[i] = s.YamlNode()
	}

	var contentNodes []*yaml.Node

	if j.Executor != "" {
		contentNodes = append(contentNodes, yScalar("executor"), yScalar(j.Executor))
	} else {
		contentNodes = append(contentNodes,
			yScalar("docker"), ySeq(yMap(
				yScalar("image"), yScalar(j.DockerImage))))
	}

	if len(j.Environment) > 0 {
		keyValPairs := make([]*yaml.Node, 2*len(j.Environment))
		i := 0
		for k, v := range j.Environment {
			keyValPairs[2*i] = yScalar(k)
			keyValPairs[2*i+1] = yScalar(v)
			i++
		}

		contentNodes = append(contentNodes, yScalar("environment"), yMap(keyValPairs...))
	}

	contentNodes = append(contentNodes, yScalar("steps"), ySeq(stepsYaml...))

	return yCommentedMap(j.Comment, contentNodes...)
}

type StepType uint32

const (
	Checkout StepType = iota
	Run
	SaveCache
	RestoreCache
	StoreArtifacts
	StoreTestResults
	OrbCommand
)

type OrbCommandParameters map[string]string

// Step definitions for Jobs. Go has no sum types, so for now just throw all supported fields under
// one struct
type Step struct {
	Type       StepType
	Comment    string
	Name       string // only used for run step
	Command    string // for run steps or orb-defined commands
	CacheKey   string
	Path       string               // cache, artifact or test results path
	Parameters OrbCommandParameters // for orb-defined steps
}

func (s Step) YamlNode() *yaml.Node {
	switch s.Type {
	case Checkout:
		return yCommentedScalar(s.Comment, "checkout")

	case Run:
		var kvs []*yaml.Node
		if s.Name != "" {
			kvs = append(kvs, yScalar("name"), yScalar(s.Name))
		}
		kvs = append(kvs, yScalar("command"), yScalar(s.Command))
		return yCommentedMap(s.Comment,
			yScalar("run"),
			yMap(kvs...))

	case SaveCache:
		// ignores Name for now
		return yCommentedMap(s.Comment,
			yScalar("save_cache"),
			yMap(
				yScalar("key"), yScalar(s.CacheKey),
				// Only support one path for now
				yScalar("paths"), ySeq(yScalar(s.Path))))

	case RestoreCache:
		// ignores Name for now
		return yCommentedMap(s.Comment,
			yScalar("restore_cache"),
			yMap(
				yScalar("key"), yScalar(s.CacheKey)))

	case StoreArtifacts:
		return yCommentedMap(s.Comment,
			yScalar("store_artifacts"),
			yMap(
				yScalar("path"), yScalar(s.Path)))

	case StoreTestResults:
		return yCommentedMap(s.Comment,
			yScalar("store_test_results"),
			yMap(
				yScalar("path"), yScalar(s.Path)))
	case OrbCommand:
		numParams := len(s.Parameters)
		if numParams == 0 {
			return yCommentedScalar(s.Comment, s.Command)
		}
		// we want the parameters of the orb command to be sorted by key for consistency
		// make an array of parameter keys
		orbParameterKeys := make([]string, numParams)
		i := 0
		for k := range s.Parameters {
			orbParameterKeys[i] = k
			i++
		}
		// then sort it
		sort.Strings(orbParameterKeys)
		// and use it to populate an array yaml scalars representing key, value, key, value...
		orbCommandYaml := make([]*yaml.Node, 2*numParams)
		for j, k := range orbParameterKeys {
			orbCommandYaml[2*j] = yScalar(k)
			orbCommandYaml[2*j+1] = yScalar(s.Parameters[k])
		}

		return yCommentedMap(s.Comment,
			yScalar(s.Command),
			yMap(orbCommandYaml...))
	}
	panic("unknown step type")
}

// helper functions to generate YAML nodes and make the above code a bit more succinct

func yScalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value}
}

func yCommentedScalar(comment string, value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, HeadComment: comment, Value: value}
}

func ySeq(items ...*yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Content: items}
}

func yMap(keyValuePairs ...*yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Content: keyValuePairs}
}

func yCommentedMap(comment string, keyValuePairs ...*yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, HeadComment: comment, Content: keyValuePairs}
}

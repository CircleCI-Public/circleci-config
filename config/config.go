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
	return yamlNodeToString(c.YamlNode())
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
	var workflowJobsYaml []*yaml.Node
	commentedOutJobs := []*yaml.Node{}
	for _, j := range w.Jobs {
		// Somewhat hacky:
		// If the WorkflowJob is marked as commented out...
		if j.CommentedOut {
			// we add it (in yaml form) to this array...
			commentedOutJobs = append(commentedOutJobs, j.YamlNode())
		} else {
			node := j.YamlNode()
			// then at the next opportunity we add a head comment with the "pending"
			// commented out jobs...
			if len(commentedOutJobs) != 0 {
				node.HeadComment = yamlNodeToString(ySeq(commentedOutJobs...))
			}
			commentedOutJobs = []*yaml.Node{}
			workflowJobsYaml = append(workflowJobsYaml, node)
		}
	}

	// if there are remaining commented out jobs we add them directly to the "jobs" scalar
	// the indentation will be inconsistent with the comments added above, but that's the best
	// we can do, I think
	jobsFootComment := ""
	if len(commentedOutJobs) != 0 {
		jobsFootComment = yamlNodeToString(ySeq(commentedOutJobs...))
	}

	return yMap(&yaml.Node{
		Kind:        yaml.ScalarNode,
		Value:       "jobs",
		FootComment: jobsFootComment,
	}, ySeq(workflowJobsYaml...))
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
	Job          *Job
	Requires     []*Job
	CommentedOut bool
}

func (wj WorkflowJob) String() string {
	return yamlNodeToString(ySeq(wj.YamlNode()))
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
	DockerImages     []string
	Executor         string
	WorkingDirectory string
	Steps            []Step
	Environment      map[string]string
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
		imageNodes := make([]*yaml.Node, 0, len(j.DockerImages))
		for _, img := range j.DockerImages {
			imageNodes = append(imageNodes, yMap(yScalar("image"), yScalar(img)))
		}
		contentNodes = append(contentNodes, yScalar("docker"), ySeq(imageNodes...))
	}

	if j.WorkingDirectory != "" && j.WorkingDirectory != "." {
		contentNodes = append(contentNodes, yScalar("working_directory"), yScalar(j.WorkingDirectory))
	}

	if len(j.Environment) > 0 {
		contentNodes = append(contentNodes,
			yScalar("environment"), yMapFromStringsMap(j.Environment))
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
		if s.Path == "" {
			return yCommentedScalar(s.Comment, "checkout")
		} else {
			return yCommentedMap(s.Comment, yScalar("checkout"),
				yMap(yScalar("path"), yScalar(s.Path)))
		}

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
		if len(s.Parameters) == 0 {
			return yCommentedScalar(s.Comment, s.Command)
		}
		return yCommentedMap(s.Comment,
			yScalar(s.Command),
			yMapFromStringsMap(s.Parameters))
	}
	panic("unknown step type")
}

func yamlNodeToString(y *yaml.Node) string {
	buf := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	err := encoder.Encode(y)
	if err != nil {
		return fmt.Sprintf("[Could not encode node: %v])", err)
	}
	return buf.String()
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

func yMapFromStringsMap(m map[string]string) *yaml.Node {
	// we want the resulting map to be sorted by key for consistency:
	// 1. make an array of keys
	numKeys := len(m)
	mapKeys := make([]string, numKeys)
	i := 0
	for k := range m {
		mapKeys[i] = k
		i++
	}
	// 2. then sort it
	sort.Strings(mapKeys)
	// 3. use it to populate an array yaml scalars representing key, value, key, value...
	contentYaml := make([]*yaml.Node, 2*numKeys)
	for j, k := range mapKeys {
		contentYaml[2*j] = yScalar(k)
		contentYaml[2*j+1] = yScalar(m[k])
	}
	return yMap(contentYaml...)
}

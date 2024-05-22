package plan_yaml

import "sort"

const (
	shell  TaskType = "sh"
	python TaskType = "python"
	exec   TaskType = "exec"
)

type privatePlanYaml struct {
	PackageId      string           `yaml:"packageId,omitempty"`
	Services       []*Service       `yaml:"services,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"`
	Tasks          []*Task          `yaml:"tasks,omitempty"`
}

// Service represents a service in the system.
type Service struct {
	Uuid       string                 `yaml:"uuid,omitempty"`
	Name       string                 `yaml:"name,omitempty"`
	Image      *ImageSpec             `yaml:"image,omitempty"`
	Cmd        []string               `yaml:"command,omitempty"`
	Entrypoint []string               `yaml:"entrypoint,omitempty"`
	EnvVars    []*EnvironmentVariable `yaml:"envVars,omitempty"`
	Ports      []*Port                `yaml:"ports,omitempty"`
	Files      []*FileMount           `yaml:"files,omitempty"`
}

func (s *Service) MarshalYAML() (interface{}, error) {
	sort.Slice(s.EnvVars, func(i, j int) bool {
		return s.EnvVars[i].Key < s.EnvVars[j].Key
	})
	return s, nil
}

type ImageSpec struct {
	ImageName string `yaml:"name,omitempty"`

	// for built images
	BuildContextLocator string `yaml:"buildContextLocator,omitempty"`
	TargetStage         string `yaml:"targetStage,omitempty"`

	// for images from registry
	Registry string `yaml:"registry,omitempty"`
}

// FilesArtifact represents a collection of files.
type FilesArtifact struct {
	Uuid  string   `yaml:"uuid,omitempty"`
	Name  string   `yaml:"name,omitempty"`
	Files []string `yaml:"files,omitempty"`
}

func (f *FilesArtifact) MarshalYAML() (interface{}, error) {
	sort.Slice(f.Files, func(i, j int) bool {
		return f.Files[i] < f.Files[j]
	})
	return f, nil
}

// EnvironmentVariable represents an environment variable.
type EnvironmentVariable struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// Port represents a port.
type Port struct {
	Name   string `yaml:"name,omitempty"`
	Number uint16 `yaml:"number,omitempty"`

	TransportProtocol   TransportProtocol   `yaml:"transportProtocol,omitempty"`
	ApplicationProtocol ApplicationProtocol `yaml:"applicationProtocol,omitempty"`
}

// ApplicationProtocol represents the application protocol used.
type ApplicationProtocol string

// TransportProtocol represents transport protocol used.
type TransportProtocol string

// FileMount represents a mount point for files.
type FileMount struct {
	MountPath      string           `yaml:"mountPath,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"` // TODO: support persistent directories
}

// Task represents a task to be executed.
type Task struct {
	Uuid     string           `yaml:"uuid,omitempty"`
	Name     string           `yaml:"name,omitempty"`
	TaskType TaskType         `yaml:"taskType,omitempty"`
	RunCmd   []string         `yaml:"command,omitempty"`
	Image    string           `yaml:"image,omitempty"`
	Files    []*FileMount     `yaml:"files,omitempty"`
	Store    []*FilesArtifact `yaml:"store,omitempty"`

	// only exists on shell tasks
	EnvVars []*EnvironmentVariable `yaml:"envVar,omitempty"`

	// only exists on python tasks
	PythonPackages []string `yaml:"pythonPackages,omitempty"`
	PythonArgs     []string `yaml:"pythonArgs,omitempty"`

	// service name
	ServiceName     string  `yaml:"serviceName,omitempty"`
	AcceptableCodes []int64 `yaml:"acceptableCodes,omitempty"`
}

// TaskType represents the type of task (either python or shell)
type TaskType string

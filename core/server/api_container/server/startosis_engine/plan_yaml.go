package startosis_engine

const (
	HTTP ApplicationProtocol = "HTTP"
	UDP  TransportProtocol   = "UDP"
	TCP  TransportProtocol   = "TCP"

	SHELL  TaskType = "sh"
	PYTHON TaskType = "python"
)

type PlanYaml struct {
	PackageId      string           `yaml:"packageId,omitempty"`
	Services       []*Service       `yaml:"services,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"`
	Tasks          []*Task          `yaml:"tasks,omitempty"`
}

// Service represents a service in the system.
type Service struct {
	Uuid       string                 `yaml:"uuid,omitempty"`       // done
	Name       string                 `yaml:"name,omitempty"`       // done
	Image      string                 `yaml:"image,omitempty"`      // done
	Cmd        []string               `yaml:"command,omitempty"`    // done
	Entrypoint []string               `yaml:"entrypoint,omitempty"` // done
	EnvVars    []*EnvironmentVariable `yaml:"envVars,omitempty"`    // done
	Ports      []*Port                `yaml:"ports,omitempty"`      // done
	Files      []*FileMount           `yaml:"files,omitempty"`
}

// FilesArtifact represents a collection of files.
type FilesArtifact struct {
	Uuid  string            `yaml:"uuid,omitempty"`
	Name  string            `yaml:"name,omitempty"`
	Files map[string]string `yaml:"files,omitempty"`
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
	filesArtifacts []*FilesArtifact // TODO: support persistent directories
}

// Task represents a task to be executed.
type Task struct {
	TaskType   TaskType               `yaml:"taskType,omitempty"`
	Name       string                 `yaml:"name,omitempty"`
	Command    string                 `yaml:"command,omitempty"`
	Image      string                 `yaml:"image,omitempty"`
	EnvVars    []*EnvironmentVariable `yaml:"envVar,omitempty"`
	Files      []*FileMount           `yaml:"files,omitempty"`
	Store      []string               `yaml:"store,omitempty"`
	ShouldWait bool                   `yaml:"shouldWait,omitempty"`
	Wait       string                 `yaml:"wait,omitempty"`
}

// TaskType represents the type of task.
type TaskType string

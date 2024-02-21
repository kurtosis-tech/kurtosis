package startosis_engine

const (
	HTTP ApplicationProtocol = "HTTP"
	UDP  TransportProtocol   = "UDP"
	TCP  TransportProtocol   = "TCP"

	SHELL  TaskType = "sh"
	PYTHON TaskType = "python"
)

// TODO: there's really no point in making any of these references, consider just making them copies
type PlanYaml struct {
	PackageId      string           `yaml:"packageId,omitempty"`
	Services       []*Service       `yaml:"services,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"`
	Tasks          []*Task          `yaml:"tasks,omitempty"`
}

// Service represents a service in the system.
type Service struct {
	Name       string                 `yaml:"name,omitempty"`       // done
	Uuid       string                 `yaml:"uuid,omitempty"`       // done
	Image      string                 `yaml:"image,omitempty"`      // done
	Cmd        []string               `yaml:"command,omitempty"`    // done
	Entrypoint []string               `yaml:"entrypoint,omitempty"` // done
	EnvVars    []*EnvironmentVariable `yaml:"envVars,omitempty"`    // done
	Ports      []*Port                `yaml:"ports,omitempty"`      // done
	Files      []*FileMount           `yaml:"files,omitempty"`      // done

	// TODO: support remaining fields in the ServiceConfig
}

// FilesArtifact represents a collection of files.
type FilesArtifact struct {
	Name  string   `yaml:"name,omitempty"`
	Uuid  string   `yaml:"uuid,omitempty"`
	Files []string `yaml:"files,omitempty"`
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
	Name     string           `yaml:"name,omitempty"`     // done
	Uuid     string           `yaml:"uuid"`               // done
	TaskType TaskType         `yaml:"taskType,omitempty"` // done
	RunCmd   string           `yaml:"command,omitempty"`  // done
	Image    string           `yaml:"image,omitempty"`    // done
	Files    []*FileMount     `yaml:"files,omitempty"`    // done
	Store    []*FilesArtifact `yaml:"store,omitempty"`    // done

	// only exists on SHELL tasks
	EnvVars []*EnvironmentVariable `yaml:"envVar,omitempty"` // done

	// only exists on PYTHON tasks
	PythonPackages []string `yaml:"pythonPackages,omitempty"`
	PythonArgs     []string `yaml:"pythonArgs,omitempty"`
}

// TaskType represents the type of task (either PYTHON or SHELL)
type TaskType string

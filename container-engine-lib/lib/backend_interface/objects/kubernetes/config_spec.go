package kubernetes

const (
	// WorkloadTypePod represents a Kubernetes Pod workload type
	WorkloadTypePod = "pod"

	// WorkloadTypeDeployment represents a Kubernetes Deployment workload type
	WorkloadTypeDeployment = "deployment"
)

type Config struct {
	ExtraIngressConfig *ExtraIngressConfig

	// WorkloadType controls what type of Kubernetes resource to create for a service
	// Valid values are:
	// - "" (empty string): defaults to "pod"
	// - "pod": creates a Pod in Kubernetes
	// - "deployment": creates a Deployment with a single replica for increased resilience
	WorkloadType string
}

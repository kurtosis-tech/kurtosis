package services

import (
	"fmt"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/kubernetes"
)

// KubernetesConfig represents the Kubernetes-specific configuration for a service
type KubernetesConfig struct {
	// ExtraIngressConfigs allows setting up Kubernetes Ingress resources to expose ports to the public internet
	ExtraIngressConfig *kubernetes.ExtraIngressConfig

	// WorkloadType controls what type of Kubernetes resource to create for a service
	// Valid values are "" (empty string, defaults to "pod"), "pod", or "deployment"
	// When set to "deployment", creates a Deployment with a single replica for increased resilience
	WorkloadType string
}

// ToStarlark converts the KubernetesConfig to its Starlark representation
func (k *KubernetesConfig) ToStarlark() string {
	if k == nil {
		return "None"
	}

	starlarkFields := []string{}

	if k.ExtraIngressConfig != nil {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`extra_ingress_config=%s`, ExtraIngressConfigToStarlark(k.ExtraIngressConfig)))
	}

	if k.WorkloadType != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`workload_type="%s"`, k.WorkloadType))
	}

	return fmt.Sprintf("KubernetesConfig(%s)", strings.Join(starlarkFields, ","))
}

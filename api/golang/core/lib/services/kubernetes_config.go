package services

import (
	"fmt"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/kubernetes"
)

// KubernetesConfig represents the Kubernetes-specific configuration for a service
type KubernetesConfig struct {
	ExtraIngressConfig *kubernetes.ExtraIngressConfig
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

	return fmt.Sprintf("KubernetesConfig(%s)", strings.Join(starlarkFields, ","))
}

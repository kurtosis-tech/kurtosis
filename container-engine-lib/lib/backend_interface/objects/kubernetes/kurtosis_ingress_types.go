package kubernetes

import (
	netv1 "k8s.io/api/networking/v1"
	"strings"
)

type TlsConfig struct {
	SecretName string
}

type PortConfig struct {
	Name   string
	Number int32
}

type HttpRule struct {
	PortConfig *PortConfig
	Path       string
	PathType   string
}

type Annotations = map[string]string

type IngressSpec struct {
	IngressClassName *string
	Host             *string
	IngressName      *string           // Todo: support with starlark
	IngressLabels    map[string]string // Todo: support with starlark
	TlsConfig        *TlsConfig
	Annotations      *Annotations
	HttpRules        []*HttpRule
}

type ExtraIngressConfig struct {
	IngressSpecs []*IngressSpec
}

func (httpRule *HttpRule) ToKubernetesHttpIngressPath(serviceName string) netv1.HTTPIngressPath {
	pt := netv1.PathType(httpRule.PathType)
	return netv1.HTTPIngressPath{
		Path:     httpRule.Path,
		PathType: &pt,
		Backend: netv1.IngressBackend{
			Service: &netv1.IngressServiceBackend{
				Name: serviceName,
				Port: netv1.ServiceBackendPort{
					Name:   httpRule.PortConfig.Name,
					Number: httpRule.PortConfig.Number,
				},
			},
		},
	}
}

func (ingressSpec *IngressSpec) GetAllKubernetesHttpIngressPaths(
	serviceName string,
) []netv1.HTTPIngressPath {

	var ingressPaths []netv1.HTTPIngressPath
	if ingressSpec.HttpRules == nil || len(ingressSpec.HttpRules) == 0 {
		return ingressPaths
	}
	for _, httpRule := range ingressSpec.HttpRules {
		ingressPaths = append(ingressPaths, httpRule.ToKubernetesHttpIngressPath(serviceName))
	}
	return ingressPaths
}

// GetKubernetesIngressRule returns the ingress rule for this ingress spec.
// We avoid returning an array here due to a design decision to only support on host per
// ingress. This simplifies the complexity of adding support for IngressController
// specific non-http ingresses should they be desired later.
func (ingressSpec *IngressSpec) GetKubernetesIngressRule(serviceName string) netv1.IngressRule {
	paths := ingressSpec.GetAllKubernetesHttpIngressPaths(serviceName)

	return netv1.IngressRule{
		Host: ingressSpec.GetHost(),
		IngressRuleValue: netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	}
}

func (ingressSpec *IngressSpec) ConstructIngressName(
	kurtosisServiceName string,
	extraIngressIdentifierSuffixOverride *string,
) string {
	var nameParts []string
	if ingressSpec.IngressClassName != nil {
		nameParts = append(nameParts, *ingressSpec.IngressClassName)
	}
	//if ingressSpec.IngressName != nil {
	//	nameParts = append(nameParts, *ingressSpec.IngressName)
	//}
	nameParts = append(nameParts, kurtosisServiceName)

	if extraIngressIdentifierSuffixOverride != nil {
		nameParts = append(nameParts, *extraIngressIdentifierSuffixOverride)
	} else {
		nameParts = append(nameParts, "user")
	}

	return strings.Join(nameParts, "-")
}

func (ingressSpec *IngressSpec) GetHost() string {
	if ingressSpec == nil {
		return ""
	}
	return *ingressSpec.Host
}

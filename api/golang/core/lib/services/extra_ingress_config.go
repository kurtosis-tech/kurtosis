package services

import (
	"fmt"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/kubernetes"
)

// ToStarlark converts the ExtraIngressConfig to its Starlark representation
func ExtraIngressConfigToStarlark(e *kubernetes.ExtraIngressConfig) string {
	if e == nil {
		return "None"
	}

	ingressStrings := []string{}
	for _, ingress := range e.IngressSpecs {
		ingressStrings = append(ingressStrings, ingressToStarlark(ingress))
	}

	return fmt.Sprintf("ExtraIngressConfig(ingresses=[%s])", strings.Join(ingressStrings, ", "))
}

// ingressToStarlark converts an IngressSpec to its Starlark representation
func ingressToStarlark(i *kubernetes.IngressSpec) string {
	if i == nil {
		return "None"
	}

	starlarkFields := []string{}

	if i.Host != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`host=%q`, i.Host))
	}

	if i.IngressClassName != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`ingress_class_name=%q`, i.IngressClassName))
	}

	if i.Annotations != nil && len(*i.Annotations) > 0 {
		annotationStrings := []string{}
		for key, value := range *i.Annotations {
			annotationStrings = append(annotationStrings, fmt.Sprintf("%q:%q", key, value))
		}
		starlarkFields = append(starlarkFields, fmt.Sprintf(`annotations={%s}`, strings.Join(annotationStrings, ",")))
	}

	if i.TlsConfig != nil {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`tls=IngressTLSConfig(secret_name=%q)`, i.TlsConfig.SecretName))
	}

	if len(i.HttpRules) > 0 {
		ruleStrings := []string{}
		for _, rule := range i.HttpRules {
			ruleStrings = append(ruleStrings, httpRuleToStarlark(rule))
		}
		starlarkFields = append(starlarkFields, fmt.Sprintf(`http_rules=[%s]`, strings.Join(ruleStrings, ", ")))
	}

	return fmt.Sprintf("IngressSpec(%s)", strings.Join(starlarkFields, ","))
}

// httpRuleToStarlark converts an HttpRule to its Starlark representation
func httpRuleToStarlark(r *kubernetes.HttpRule) string {
	if r == nil {
		return "None"
	}

	starlarkFields := []string{}

	if r.Path != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`path=%q`, r.Path))
	}

	if r.PathType != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`path_type=%q`, r.PathType))
	}

	if r.PortConfig != nil {
		if r.PortConfig.Name != "" {
			starlarkFields = append(starlarkFields, fmt.Sprintf(`port_name=%q`, r.PortConfig.Name))
		}
		if r.PortConfig.Number != 0 {
			starlarkFields = append(starlarkFields, fmt.Sprintf(`port=%d`, r.PortConfig.Number))
		}
	}

	return fmt.Sprintf("IngressHttpRule(%s)", strings.Join(starlarkFields, ","))
}

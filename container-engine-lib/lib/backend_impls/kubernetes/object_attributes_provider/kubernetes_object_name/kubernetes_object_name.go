package kubernetes_object_name

import (
	"github.com/kurtosis-tech/stacktrace"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Represents a Kubernetes label that is guaranteed to be valid for the kubernetes cluster
// NOTE: This is a struct-based enum
type KubernetesObjectName struct {
	value string
}

// NOTE: This is ONLY for areas where the label value is declared statically!! Any sort of dynamic/runtime label value creation
//  should use CreateNewKubernetesObjectName
func MustCreateNewKubernetesObjectName(str string) *KubernetesObjectName {
	name, err := CreateNewKubernetesObjectName(str)
	if err != nil {
		panic(err)
	}
	return name
}

func CreateNewKubernetesObjectName(str string) (*KubernetesObjectName, error) {
	if err := validateKubernetesObjectName(str); err != nil {
		return nil, stacktrace.Propagate(err, "Object name string '%v' doesn't pass validation of being a kubernetes object name", str)
	}

	return &KubernetesObjectName{value: str}, nil
}
func (key *KubernetesObjectName) GetString() string {
	return key.value
}

// https://github.com/kubernetes/design-proposals-archive/blob/main/architecture/identifiers.md
// In kubernetes, to create an object you must specify a 'name' that is a DNS_LABEL, following rfc1123
func validateKubernetesObjectName(str string) error {
	validationErrs := validation.IsDNS1123Label(str)
	if len(validationErrs) > 0 {
		// TODO pretty up printing of errors
		// return err
		return stacktrace.NewError("Expected object name string '%v' to be a valid DNS_LABEL, instead it failed validation:\n%+v", str, validationErrs)
	}
	return nil
}

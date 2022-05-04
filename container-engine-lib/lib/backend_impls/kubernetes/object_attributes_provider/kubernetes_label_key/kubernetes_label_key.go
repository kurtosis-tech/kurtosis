package kubernetes_label_key

import (
	"github.com/kurtosis-tech/stacktrace"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

// Represents a Kubernetes label ney that is guaranteed to be valid for Kubernetes
type KubernetesLabelKey struct {
	value string
}

// NOTE: This is ONLY for areas where the label is declared statically!! Any sort of dynamic/runtime label creation
//  should use CreateNewKubernetesLabelKey
func MustCreateNewKubernetesLabelKey(str string) *KubernetesLabelKey {
	key, err := CreateNewKubernetesLabelKey(str)
	if err != nil {
		panic(err)
	}
	return key
}
func CreateNewKubernetesLabelKey(str string) (*KubernetesLabelKey, error) {
	if err := validateLabelKey(str); err != nil {
		return nil, stacktrace.Propagate(err, "Label value string '%v' doesn't pass validation of being a kubernetes label key", str)
	}

	return &KubernetesLabelKey{value: str}, nil
}
func (key *KubernetesLabelKey) GetString() string {
	return key.value
}

func validateLabelKey(str string) error {
	validationErrs := validation.IsQualifiedName(str)
	if len(validationErrs) > 0 {
		errString := strings.Join(validationErrs, "\n\n")
		return stacktrace.NewError("Expected label string '%v' to be a valid kubernetes label key, instead it failed validation:\n%+v", str, errString)
	}
	return nil

}

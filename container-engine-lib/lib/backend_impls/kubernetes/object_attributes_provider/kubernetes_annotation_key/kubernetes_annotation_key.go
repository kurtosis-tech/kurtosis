package kubernetes_annotation_key

import (
	"github.com/kurtosis-tech/stacktrace"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

// Represents a Kubernetes label ney that is guaranteed to be valid for Kubernetes
type KubernetesAnnotationKey struct {
	value string
}

// NOTE: This is ONLY for areas where the label is declared statically!! Any sort of dynamic/runtime label creation
//  should use CreateNewKubernetesLabelKey
func MustCreateNewKubernetesAnnotationKey(str string) *KubernetesAnnotationKey {
	key, err := CreateNewKubernetesAnnotationKey(str)
	if err != nil {
		panic(err)
	}
	return key
}

func CreateNewKubernetesAnnotationKey(str string) (*KubernetesAnnotationKey, error) {
	if err := validateAnnotationKey(str); err != nil {
		return nil, stacktrace.Propagate(err, "Annotation value string '%v' doesn't pass validation of being a kubernetes annotation key", str)
	}

	return &KubernetesAnnotationKey{value: str}, nil
}
func (key *KubernetesAnnotationKey) GetString() string {
	return key.value
}

func validateAnnotationKey(str string) error {
	validationErrs := validation.IsQualifiedName(str)
	if len(validationErrs) > 0 {
		errString := strings.Join(validationErrs, "\n\n")
		return stacktrace.NewError("Expected string '%v' to be a valid kubernetes annotation key, instead it failed validation:\n%+v", str, errString)
	}
	return nil

}

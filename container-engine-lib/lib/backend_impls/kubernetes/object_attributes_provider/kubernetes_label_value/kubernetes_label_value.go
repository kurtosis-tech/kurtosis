package kubernetes_label_value

import (
	"github.com/kurtosis-tech/stacktrace"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

// Represents a Kubernetes label value that is guaranteed to be valid for kubernetes
// NOTE: This is a struct-based enum
type KubernetesLabelValue struct {
	value string
}

// NOTE: This is ONLY for areas where the label value is declared statically!! Any sort of dynamic/runtime label value creation
//  should use CreateNewKubernetesLabelValue
func MustCreateNewKubernetesLabelValue(str string) *KubernetesLabelValue {
	key, err := CreateNewKubernetesLabelValue(str)
	if err != nil {
		panic(err)
	}
	return key
}
func CreateNewKubernetesLabelValue(labelValue string) (*KubernetesLabelValue, error) {
	if err := validateLabelValue(labelValue); err != nil {
		return nil, stacktrace.Propagate(err, "Label value string '%v' doesn't pass validation of being a kubernetes label value", labelValue)
	}

	return &KubernetesLabelValue{value: labelValue}, nil
}
func (key *KubernetesLabelValue) GetString() string {

	return key.value
}

// validateLabelStr throws an error if str isn't a "qualified name" in kubernetes
func validateLabelValue(str string) error {
	validationErrs := validation.IsValidLabelValue(str)
	if len(validationErrs) > 0 {
		errString := strings.Join(validationErrs, "\n\n")
		return stacktrace.NewError("Expected label string '%v' to be a kubernetes label value, instead it failed validation:\n%+v", str, errString)
	}
	return nil
}

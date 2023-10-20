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
// should use createNewKubernetesLabelKey
func MustCreateNewKubernetesLabelKey(str string) *KubernetesLabelKey {
	key, err := createNewKubernetesLabelKey(str)
	if err != nil {
		panic(err)
	}
	return key
}

// CreateNewKubernetesUserCustomLabelKey creates a custom uer Kubernetes label with the Kurtosis custom user prefix
func CreateNewKubernetesUserCustomLabelKey(str string) (*KubernetesLabelKey, error) {
	if str == "" || str == " " {
		return nil, stacktrace.NewError("Received an empty user custom label key")
	}
	labelKeyStr := customUserLabelsKeyPrefixStr + str
	return createNewKubernetesLabelKey(labelKeyStr)
}

func createNewKubernetesLabelKey(str string) (*KubernetesLabelKey, error) {
	if err := validateLabelKey(str); err != nil {
		return nil, stacktrace.Propagate(err, "Label value string '%v' doesn't pass validation of being a Kubernetes label key", str)
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
		return stacktrace.NewError("Expected label string '%v' to be a valid Kubernetes label key, instead it failed validation:\n%+v", str, errString)
	}
	return nil

}

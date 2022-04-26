package kubernetes_annotation_value

// Represents a Kubernetes label value that is guaranteed to be valid for the Kubernetes cluster
// NOTE: This is a struct-based enum
type KubernetesAnnotationValue struct {
	value string
}

func CreateNewKubernetesAnnotationValue(str string) (*KubernetesAnnotationValue, error) {
	// In k8s, an annotation value is an arbitrary string of bytes
	return &KubernetesAnnotationValue{value: str}, nil
}

func (key *KubernetesAnnotationValue) GetString() string {
	return key.value
}

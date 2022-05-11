package kubernetes_resource_collectors

type kubernetesResource interface {
	getName() string
	getLabels() map[string]string
	getUnderlying() interface{}	// Used when we downcast back, because we don't have generics yet
}

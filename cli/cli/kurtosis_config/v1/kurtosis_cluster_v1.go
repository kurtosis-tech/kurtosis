package v1

type KurtosisClusterV1 struct {
	Type *string                      `yaml:"type,omitempty"`
	// If we ever get another type of cluster that has configuration, this will need to be polymorphically deserialized
	Config *KubernetesClusterConfigV1 `yaml:"config,omitempty"`
}

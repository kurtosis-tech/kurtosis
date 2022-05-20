package v2

type KurtosisClusterConfigV2 struct {
	Type *string                      `yaml:"type,omitempty"`
	// If we ever get another type of cluster that has configuration, this will need to be polymorphically deserialized
	Config *KubernetesClusterConfigV2 `yaml:"config,omitempty"`
}

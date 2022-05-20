package v2

type KubernetesClusterConfigV2 struct {
	KubernetesClusterName *string `yaml:"kubernetes-cluster-name,omitempty"`
	StorageClass *string `yaml:"storage-class,omitempty"`
	EnclaveSizeInMegabytes *uint `yaml:"enclave-size-in-megabytes,omitempty"`
}

